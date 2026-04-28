package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"sanctum/internal/models"
	"sanctum/internal/notifications"
	"sanctum/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockChatRepository is a mock of the ChatRepository interface
type MockChatRepository struct {
	mock.Mock
}

func TestSendMessagePublishesExactlyOneRedisChatMessage(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()

	ctx := context.Background()
	sub := rdb.Subscribe(ctx, "chat:conv:1")
	defer func() { _ = sub.Close() }()
	_, err = sub.Receive(ctx)
	require.NoError(t, err)

	app := fiber.New()
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)
	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil, nil, nil)
	s := &Server{
		chatRepo:    mockChatRepo,
		userRepo:    mockUserRepo,
		chatService: chatService,
		notifier:    notifications.NewNotifier(rdb),
	}

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", uint(1))
		return c.Next()
	})
	app.Post("/conversations/:id/messages", s.SendMessage)

	conv := &models.Conversation{
		ID:        1,
		IsGroup:   true,
		CreatedBy: 1,
		Participants: []models.User{
			{ID: 1, Username: "sender"},
			{ID: 2, Username: "reader"},
		},
	}
	mockChatRepo.On("GetConversation", mock.Anything, uint(1)).Return(conv, nil).Once()
	mockChatRepo.On("CreateMessage", mock.Anything, mock.MatchedBy(func(msg *models.Message) bool {
		msg.ID = 10
		return msg.ConversationID == 1 && msg.SenderID == 1 && msg.Content == "hello"
	})).Return(nil).Once()
	mockUserRepo.On("GetByID", mock.Anything, uint(1)).
		Return(&models.User{ID: 1, Username: "sender"}, nil).Once()

	body := bytes.NewReader([]byte(`{"content":"hello"}`))
	req := newRequest(http.MethodPost, "/conversations/1/messages", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	msg, err := sub.ReceiveMessage(ctx)
	require.NoError(t, err)
	assert.Equal(t, "chat:conv:1", msg.Channel)

	select {
	case extra := <-sub.Channel():
		t.Fatalf("unexpected extra chat publish: %s", extra.Payload)
	case <-time.After(50 * time.Millisecond):
	}

	mockChatRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func (m *MockChatRepository) CreateConversation(ctx context.Context, conv *models.Conversation) error {
	args := m.Called(ctx, conv)
	return args.Error(0)
}

func (m *MockChatRepository) GetConversation(ctx context.Context, id uint) (*models.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Conversation), args.Error(1)
}

func (m *MockChatRepository) GetUserConversations(ctx context.Context, userID uint) ([]*models.Conversation, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Conversation), args.Error(1)
}

func (m *MockChatRepository) AddParticipant(ctx context.Context, convID, userID uint) error {
	args := m.Called(ctx, convID, userID)
	return args.Error(0)
}

func (m *MockChatRepository) RemoveParticipant(ctx context.Context, convID, userID uint) error {
	args := m.Called(ctx, convID, userID)
	return args.Error(0)
}

func (m *MockChatRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockChatRepository) GetMessages(ctx context.Context, convID uint, limit, offset int) ([]*models.Message, error) {
	args := m.Called(ctx, convID, limit, offset)
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockChatRepository) MarkMessageRead(ctx context.Context, msgID uint) error {
	args := m.Called(ctx, msgID)
	return args.Error(0)
}

func (m *MockChatRepository) UpdateLastRead(ctx context.Context, convID, userID uint) error {
	args := m.Called(ctx, convID, userID)
	return args.Error(0)
}

func (m *MockChatRepository) IsUserParticipant(ctx context.Context, conversationID, userID uint) (bool, error) {
	args := m.Called(ctx, conversationID, userID)
	return args.Bool(0), args.Error(1)
}

func TestCreateConversation(t *testing.T) {
	app := fiber.New()
	mockChatRepo := new(MockChatRepository)
	chatService := service.NewChatService(mockChatRepo, nil, nil, nil, nil)
	s := &Server{chatRepo: mockChatRepo, chatService: chatService}

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", uint(1))
		return c.Next()
	})
	app.Post("/conversations", s.CreateConversation)

	tests := []struct {
		name           string
		body           map[string]interface{}
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Success",
			body: map[string]interface{}{
				"participant_ids": []uint{2},
			},
			mockSetup: func() {
				mockChatRepo.On("CreateConversation", mock.Anything, mock.Anything).Return(nil)
				mockChatRepo.On("AddParticipant", mock.Anything, mock.Anything, uint(1)).Return(nil)
				mockChatRepo.On("AddParticipant", mock.Anything, mock.Anything, uint(2)).Return(nil)
				mockChatRepo.On("GetConversation", mock.Anything, mock.Anything).Return(&models.Conversation{ID: 1}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing Participants",
			body: map[string]interface{}{
				"participant_ids": []uint{},
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			body, err := json.Marshal(tt.body)
			assert.NoError(t, err)
			req := newRequest(http.MethodPost, "/conversations", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, _ := app.Test(req)
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

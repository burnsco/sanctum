package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"sanctum/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetUserProfile(t *testing.T) {
	app := fiber.New()
	mockRepo := new(MockUserRepository)
	s := &Server{userRepo: mockRepo}

	app.Get("/users/:id", s.GetUserProfile)

	tests := []struct {
		name           string
		userIDParam    string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:        "Success",
			userIDParam: "1",
			mockSetup: func() {
				mockRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			userIDParam:    "abc",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Not Found",
			userIDParam: "99",
			mockSetup: func() {
				mockRepo.On("GetByID", mock.Anything, uint(99)).Return(nil, models.NewNotFoundError("User", 99))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := newRequest(http.MethodGet, "/users/"+tt.userIDParam, nil)
			resp, _ := app.Test(req)
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestGetUserProfile_SanitizesSensitiveFields(t *testing.T) {
	app := fiber.New()
	mockRepo := new(MockUserRepository)
	s := &Server{userRepo: mockRepo}

	app.Get("/users/:id", s.GetUserProfile)

	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.User{
		ID:                1,
		Username:          "testuser",
		Email:             "test@example.com",
		IsAdmin:           true,
		IsBanned:          true,
		BannedReason:      "spam",
		ModerationStrikes: 4,
	}, nil).Once()

	req := newRequest(http.MethodGet, "/users/1", nil)
	resp, _ := app.Test(req)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var user models.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	assert.Empty(t, user.Email)
	assert.False(t, user.IsAdmin)
	assert.False(t, user.IsBanned)
	assert.Empty(t, user.BannedReason)
	assert.Zero(t, user.ModerationStrikes)
}

func TestGetMyProfile(t *testing.T) {
	app := fiber.New()
	mockRepo := new(MockUserRepository)
	s := &Server{userRepo: mockRepo}

	// Middleware to set userID in Locals
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", uint(1))
		return c.Next()
	})
	app.Get("/users/me", s.GetMyProfile)

	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(&models.User{
		ID:       1,
		Username: "me",
		Email:    "me@example.com",
		IsBanned: true,
	}, nil)

	req := newRequest(http.MethodGet, "/users/me", nil)
	resp, _ := app.Test(req)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var user models.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	assert.Equal(t, "me@example.com", user.Email)
	assert.True(t, user.IsBanned)
}

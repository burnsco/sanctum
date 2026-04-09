package server

import (
	"testing"

	"sanctum/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeSharedConversation(t *testing.T) {
	t.Parallel()

	conversation := &models.Conversation{
		ID:      7,
		IsGroup: true,
		Name:    "room",
		SanctumID: func() *uint {
			id := uint(3)
			return &id
		}(),
		Participants: []models.User{
			{ID: 1, Username: "one", Email: "one@example.com", IsBanned: true, ModerationStrikes: 2},
		},
		Messages: []models.Message{
			{
				ID:       9,
				SenderID: 2,
				Sender: &models.User{
					ID:           2,
					Username:     "two",
					Email:        "two@example.com",
					BannedReason: "abuse",
				},
			},
		},
	}

	sanitizeSharedConversation(conversation)

	require.Len(t, conversation.Participants, 1)
	assert.Empty(t, conversation.Participants[0].Email)
	assert.False(t, conversation.Participants[0].IsBanned)
	assert.Zero(t, conversation.Participants[0].ModerationStrikes)

	require.Len(t, conversation.Messages, 1)
	require.NotNil(t, conversation.Messages[0].Sender)
	assert.Empty(t, conversation.Messages[0].Sender.Email)
	assert.Empty(t, conversation.Messages[0].Sender.BannedReason)
}

func TestSanitizeSharedFriendship(t *testing.T) {
	t.Parallel()

	friendship := &models.Friendship{
		Requester: models.User{
			ID:                1,
			Username:          "requester",
			Email:             "requester@example.com",
			ModerationStrikes: 3,
		},
		Addressee: models.User{
			ID:           2,
			Username:     "addressee",
			Email:        "addressee@example.com",
			BannedReason: "spam",
		},
	}

	sanitizeSharedFriendship(friendship)

	assert.Empty(t, friendship.Requester.Email)
	assert.Zero(t, friendship.Requester.ModerationStrikes)
	assert.Empty(t, friendship.Addressee.Email)
	assert.Empty(t, friendship.Addressee.BannedReason)
}

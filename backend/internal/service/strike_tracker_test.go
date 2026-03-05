package service

import (
	"context"
	"testing"

	"sanctum/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newStrikeTrackerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}))
	return db
}

func TestGORMStrikeTracker_RecordStrike_Increments(t *testing.T) {
	db := newStrikeTrackerDB(t)
	u := &models.User{Username: "testuser", Email: "u@e.com"}
	db.Create(u)

	tracker := NewGORMStrikeTracker(db)
	ctx := context.Background()

	strikes, isBanned, err := tracker.RecordStrike(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, strikes)
	assert.False(t, isBanned)

	strikes, isBanned, err = tracker.RecordStrike(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, strikes)
	assert.False(t, isBanned)
}

func TestGORMStrikeTracker_RecordStrike_AutoBansAtLimit(t *testing.T) {
	db := newStrikeTrackerDB(t)
	u := &models.User{Username: "badactor", Email: "bad@e.com", ModerationStrikes: 2}
	db.Create(u)

	tracker := NewGORMStrikeTracker(db)
	ctx := context.Background()

	strikes, isBanned, err := tracker.RecordStrike(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, strikes)
	assert.True(t, isBanned)

	// Confirm ban persisted in DB.
	var saved models.User
	db.First(&saved, u.ID)
	assert.True(t, saved.IsBanned)
	assert.NotNil(t, saved.BannedAt)
	assert.NotEmpty(t, saved.BannedReason)
}

func TestGORMStrikeTracker_RecordStrike_AlreadyBannedNoDoubleBan(t *testing.T) {
	db := newStrikeTrackerDB(t)
	// User already at limit and already banned.
	u := &models.User{Username: "already", Email: "already@e.com", ModerationStrikes: 3, IsBanned: true}
	db.Create(u)

	tracker := NewGORMStrikeTracker(db)
	ctx := context.Background()

	strikes, isBanned, err := tracker.RecordStrike(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, strikes)
	// isBanned is true because the user was already banned (returned as true).
	assert.True(t, isBanned)
}

func TestGORMStrikeTracker_RecordStrike_UserNotFound(t *testing.T) {
	db := newStrikeTrackerDB(t)
	tracker := NewGORMStrikeTracker(db)

	_, _, err := tracker.RecordStrike(context.Background(), 9999)
	assert.Error(t, err)
}

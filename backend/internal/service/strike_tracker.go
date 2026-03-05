package service

import (
	"context"
	"time"

	"sanctum/internal/cache"
	"sanctum/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const moderationStrikeLimit = 3

// StrikeTracker records automated content moderation violations for a user
// and auto-bans the account when the strike limit is reached.
type StrikeTracker interface {
	RecordStrike(ctx context.Context, userID uint) (strikes int, isBanned bool, err error)
}

type gormStrikeTracker struct {
	db *gorm.DB
}

// NewGORMStrikeTracker returns a StrikeTracker backed by a GORM database.
func NewGORMStrikeTracker(db *gorm.DB) StrikeTracker {
	return &gormStrikeTracker{db: db}
}

// RecordStrike increments the user's moderation strike count.
// If the count reaches moderationStrikeLimit the account is auto-banned.
// Returns the new strike count and whether the ban was applied.
func (t *gormStrikeTracker) RecordStrike(ctx context.Context, userID uint) (int, bool, error) {
	var user models.User

	err := t.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, userID).Error
	if err != nil {
		return 0, false, err
	}

	user.ModerationStrikes++
	isBanned := false
	if user.ModerationStrikes >= moderationStrikeLimit && !user.IsBanned {
		user.IsBanned = true
		now := time.Now()
		user.BannedAt = &now
		user.BannedReason = "Automated ban: repeated content policy violations"
		isBanned = true
	}

	if err := t.db.WithContext(ctx).Save(&user).Error; err != nil {
		return 0, false, err
	}

	// Invalidate the user cache so the next auth check picks up the ban.
	cache.InvalidateUser(ctx, userID)

	return user.ModerationStrikes, isBanned || user.IsBanned, nil
}

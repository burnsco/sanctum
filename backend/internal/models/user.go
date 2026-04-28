// Package models contains data structures for the application's domain models.
package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the Sanctum application.
type User struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Username          string         `gorm:"unique;not null" json:"username"`
	Email             string         `gorm:"unique;not null" json:"email"`
	Password          string         `gorm:"not null" json:"-"`
	Bio               string         `json:"bio"`
	Avatar            string         `json:"avatar"`
	IsAdmin           bool           `gorm:"default:false" json:"is_admin"`
	IsBanned          bool           `gorm:"default:false" json:"is_banned"`
	BannedAt          *time.Time     `json:"banned_at,omitempty"`
	BannedReason      string         `gorm:"type:text;default:''" json:"banned_reason,omitempty"`
	BannedByUserID    *uint          `json:"banned_by_user_id,omitempty"`
	ModerationStrikes int            `gorm:"default:0" json:"moderation_strikes,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	Posts             []Post         `gorm:"foreignKey:UserID" json:"posts,omitempty"`
}

// UserSummary is the safe public representation of a user embedded in public responses.
type UserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// NewUserSummary returns the public-safe user DTO for shared API responses.
func NewUserSummary(user User) UserSummary {
	return UserSummary{
		ID:       user.ID,
		Username: user.Username,
		Avatar:   user.Avatar,
	}
}

// NewUserSummaries returns public-safe user DTOs for shared API responses.
func NewUserSummaries(users []User) []UserSummary {
	summaries := make([]UserSummary, 0, len(users))
	for _, user := range users {
		summaries = append(summaries, NewUserSummary(user))
	}
	return summaries
}

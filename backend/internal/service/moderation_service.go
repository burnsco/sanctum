package service

import (
	"context"
	"log/slog"
	"time"

	"sanctum/internal/models"

	"gorm.io/gorm"
)

// DeletedPostRow is a minimal row for admin deleted-posts listing (no content/title/image).
type DeletedPostRow struct {
	ID        uint       `json:"id"`
	UserID    uint       `json:"user_id"`
	Username  string     `json:"username"`
	PostType  string     `json:"post_type"`
	SanctumID *uint      `json:"sanctum_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// DeletedCommentRow is a minimal row for admin deleted-comments listing (no content).
type DeletedCommentRow struct {
	ID        uint       `json:"id"`
	UserID    uint       `json:"user_id"`
	Username  string     `json:"username"`
	PostID    uint       `json:"post_id"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// BanRequestRow is a row for admin ban-request listing.
type BanRequestRow struct {
	ReportedUserID uint        `json:"reported_user_id"`
	ReportCount    int64       `json:"report_count"`
	LatestReportAt time.Time   `json:"latest_report_at"`
	User           models.User `json:"user"`
}

// AdminUserDetail aggregates user and moderation data for admin views.
type AdminUserDetail struct {
	User           models.User               `json:"user"`
	Reports        []models.ModerationReport `json:"reports"`
	ActiveMutes    []models.ChatroomMute     `json:"active_mutes"`
	BlocksGiven    []models.UserBlock        `json:"blocks_given"`
	BlocksReceived []models.UserBlock        `json:"blocks_received"`
	Warnings       []string                  `json:"warnings,omitempty"`
}

// ModerationService provides admin moderation and reporting logic.
type ModerationService struct {
	db *gorm.DB
}

// NewModerationService returns a new ModerationService.
func NewModerationService(db *gorm.DB) *ModerationService {
	return &ModerationService{db: db}
}

// GetAdminBanRequests returns aggregated ban-request rows for admin.
func (s *ModerationService) GetAdminBanRequests(ctx context.Context, limit, offset int) ([]BanRequestRow, error) {
	type RawRow struct {
		ReportedUserID uint      `json:"reported_user_id"`
		ReportCount    int64     `json:"report_count"`
		LatestReportAt time.Time `json:"latest_report_at"`
	}

	var rows []RawRow
	if err := s.db.WithContext(ctx).
		Table("moderation_reports").
		Select("reported_user_id, COUNT(*) as report_count, MAX(created_at) as latest_report_at").
		Where("status = ? AND target_type = ? AND reported_user_id IS NOT NULL", models.ReportStatusOpen, models.ReportTargetUser).
		Group("reported_user_id").
		Order("report_count DESC, latest_report_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	userIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.ReportedUserID)
	}

	usersByID := map[uint]models.User{}
	if len(userIDs) > 0 {
		var users []models.User
		if err := s.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			usersByID[user.ID] = user
		}
	}

	resp := make([]BanRequestRow, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, BanRequestRow{
			ReportedUserID: row.ReportedUserID,
			ReportCount:    row.ReportCount,
			LatestReportAt: row.LatestReportAt,
			User:           usersByID[row.ReportedUserID],
		})
	}
	return resp, nil
}

// GetAdminUserDetail returns detailed user and moderation data for admin.
func (s *ModerationService) GetAdminUserDetail(ctx context.Context, userID uint) (*AdminUserDetail, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}

	detail := &AdminUserDetail{
		User: user,
	}

	// 1. Reports
	if err := s.db.WithContext(ctx).
		Where("reported_user_id = ?", userID).
		Order("created_at DESC").
		Limit(200).
		Find(&detail.Reports).Error; err != nil {
		slog.WarnContext(ctx, "failed to load reports for user", "user_id", userID, "err", err)
		detail.Warnings = append(detail.Warnings, "Partial data: Moderation reports could not be loaded.")
	}

	// 2. Active Mutes
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&detail.ActiveMutes).Error; err != nil {
		slog.WarnContext(ctx, "failed to load active mutes for user", "user_id", userID, "err", err)
		detail.Warnings = append(detail.Warnings, "Partial data: Active mutes could not be loaded.")
	}

	// 3. Blocks Given
	if err := s.db.WithContext(ctx).
		Where("blocker_id = ?", userID).
		Order("created_at DESC").
		Limit(200).
		Find(&detail.BlocksGiven).Error; err != nil {
		slog.WarnContext(ctx, "failed to load blocks given for user", "user_id", userID, "err", err)
		detail.Warnings = append(detail.Warnings, "Partial data: Outgoing blocks could not be loaded.")
	}

	// 4. Blocks Received
	if err := s.db.WithContext(ctx).
		Where("blocked_id = ?", userID).
		Order("created_at DESC").
		Limit(200).
		Find(&detail.BlocksReceived).Error; err != nil {
		slog.WarnContext(ctx, "failed to load blocks received for user", "user_id", userID, "err", err)
		detail.Warnings = append(detail.Warnings, "Partial data: Incoming blocks could not be loaded.")
	}

	return detail, nil
}

// GetAdminDeletedPosts returns deleted posts (soft-deleted) for admin listing.
func (s *ModerationService) GetAdminDeletedPosts(ctx context.Context, limit, offset int) ([]DeletedPostRow, error) {
	type rawPost struct {
		ID        uint       `gorm:"column:id"`
		UserID    uint       `gorm:"column:user_id"`
		Username  string     `gorm:"column:username"`
		PostType  string     `gorm:"column:post_type"`
		SanctumID *uint      `gorm:"column:sanctum_id"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
	}
	var rows []rawPost
	if err := s.db.WithContext(ctx).
		Unscoped().
		Table("posts").
		Select("posts.id, posts.user_id, users.username, posts.post_type, posts.sanctum_id, posts.created_at, posts.deleted_at").
		Joins("LEFT JOIN users ON users.id = posts.user_id").
		Where("posts.deleted_at IS NOT NULL").
		Order("posts.deleted_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]DeletedPostRow, 0, len(rows))
	for _, r := range rows {
		result = append(result, DeletedPostRow(r))
	}
	return result, nil
}

// GetAdminDeletedComments returns deleted comments (soft-deleted) for admin listing.
func (s *ModerationService) GetAdminDeletedComments(ctx context.Context, limit, offset int) ([]DeletedCommentRow, error) {
	type rawComment struct {
		ID        uint       `gorm:"column:id"`
		UserID    uint       `gorm:"column:user_id"`
		Username  string     `gorm:"column:username"`
		PostID    uint       `gorm:"column:post_id"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
	}
	var rows []rawComment
	if err := s.db.WithContext(ctx).
		Unscoped().
		Table("comments").
		Select("comments.id, comments.user_id, users.username, comments.post_id, comments.created_at, comments.deleted_at").
		Joins("LEFT JOIN users ON users.id = comments.user_id").
		Where("comments.deleted_at IS NOT NULL").
		Order("comments.deleted_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]DeletedCommentRow, 0, len(rows))
	for _, r := range rows {
		result = append(result, DeletedCommentRow(r))
	}
	return result, nil
}

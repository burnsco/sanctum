package service

import (
	"context"
	"errors"

	"sanctum/internal/models"
)

// --- Moderator stub ---

type moderatorStub struct {
	checkFn          func(ctx context.Context, text string) error
	checkWithImageFn func(ctx context.Context, text, imageURL string) error
}

func (m *moderatorStub) Check(ctx context.Context, text string) error {
	if m.checkFn != nil {
		return m.checkFn(ctx, text)
	}
	return nil
}

func (m *moderatorStub) CheckWithImage(ctx context.Context, text, imageURL string) error {
	if m.checkWithImageFn != nil {
		return m.checkWithImageFn(ctx, text, imageURL)
	}
	return nil
}

func allowAllModerator() *moderatorStub { return &moderatorStub{} }

func blockAllModerator() *moderatorStub {
	violation := errors.New("content violates community guidelines")
	return &moderatorStub{
		checkFn:          func(_ context.Context, _ string) error { return violation },
		checkWithImageFn: func(_ context.Context, _, _ string) error { return violation },
	}
}

// --- StrikeTracker stub ---

type strikeTrackerStub struct {
	recordFn func(ctx context.Context, userID uint) (int, bool, error)
}

func (s *strikeTrackerStub) RecordStrike(ctx context.Context, userID uint) (int, bool, error) {
	if s.recordFn != nil {
		return s.recordFn(ctx, userID)
	}
	return 1, false, nil
}

func fixedStrikeTracker(strikes int, isBanned bool) *strikeTrackerStub {
	return &strikeTrackerStub{
		recordFn: func(_ context.Context, _ uint) (int, bool, error) {
			return strikes, isBanned, nil
		},
	}
}

// --- assertModerationViolationError helper ---

func assertModerationViolationError(t interface {
	Helper()
	Errorf(format string, args ...interface{})
	FailNow()
}, err error, wantStrikes int, wantBanned bool) {
	type helperT interface {
		Helper()
		Errorf(format string, args ...interface{})
		FailNow()
	}

	if err == nil {
		t.Helper()
		t.Errorf("expected ModerationViolationError, got nil")
		t.FailNow()
	}
	var modErr *models.ModerationViolationError
	found := false
	// Walk the chain manually since errors.As works on interfaces.
	e := err
	for e != nil {
		if me, ok := e.(*models.ModerationViolationError); ok {
			modErr = me
			found = true
			break
		}
		e = errors.Unwrap(e)
	}
	if !found {
		t.Helper()
		t.Errorf("expected *models.ModerationViolationError, got %T: %v", err, err)
		t.FailNow()
		return
	}
	if modErr.Strikes != wantStrikes {
		t.Helper()
		t.Errorf("strikes: want %d, got %d", wantStrikes, modErr.Strikes)
	}
	if modErr.IsBanned != wantBanned {
		t.Helper()
		t.Errorf("isBanned: want %v, got %v", wantBanned, modErr.IsBanned)
	}
}

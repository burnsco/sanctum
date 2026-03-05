package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModerationViolationError(t *testing.T) {
	t.Parallel()

	err := NewModerationViolationError("bad content", 2, false)
	require.NotNil(t, err)
	assert.Equal(t, "MODERATION_VIOLATION", err.Code)
	assert.Equal(t, "bad content", err.Message)
	assert.Equal(t, 2, err.Strikes)
	assert.False(t, err.IsBanned)
}

func TestModerationViolationError_IsBanned(t *testing.T) {
	t.Parallel()

	err := NewModerationViolationError("repeated violation", 3, true)
	assert.Equal(t, 3, err.Strikes)
	assert.True(t, err.IsBanned)
}

func TestModerationViolationError_ErrorMethod(t *testing.T) {
	t.Parallel()

	err := NewModerationViolationError("policy violation", 1, false)
	assert.Equal(t, "policy violation", err.Error())
}

func TestModerationViolationError_ErrorsAs(t *testing.T) {
	t.Parallel()

	orig := NewModerationViolationError("flagged", 2, false)

	// errors.As must match *ModerationViolationError.
	var modErr *ModerationViolationError
	require.True(t, errors.As(orig, &modErr))
	assert.Equal(t, 2, modErr.Strikes)
	assert.Equal(t, "MODERATION_VIOLATION", modErr.Code)
}

func TestModerationViolationError_Unwrap(t *testing.T) {
	t.Parallel()

	err := NewModerationViolationError("msg", 1, false)
	// Unwrap exposes the inner AppError chain; it should not panic.
	_ = err.Unwrap()
}

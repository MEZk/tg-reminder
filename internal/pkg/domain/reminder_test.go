package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DefaultAttemptsLeft(t *testing.T) {
	t.Parallel()
	require.EqualValues(t, 3, DefaultAttemptsLeft)
}

func TestReminder_String(t *testing.T) {
	t.Parallel()
	reminder := Reminder{
		ID:           1,
		ChatID:       2,
		UserID:       3,
		Text:         "sagdshyjghj",
		RemindAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:       ReminderStatusPending,
		AttemptsLeft: 3,
	}
	assert.Equal(t, "[ID: 1, UserID: 3, ChatID: 2, Status: pending, RemindAt: 2020-01-01 00:00:00 +0000 UTC, AttemptsLeft: 3, Data: sagdshyjghj]", reminder.String())

	assert.EqualValues(t, "pending", ReminderStatusPending)
	assert.EqualValues(t, "attempts_exhausted", ReminderStatusAttemptsExhausted)
	assert.EqualValues(t, "done", ReminderStatusDone)
}

func TestReminder_UserFormat(t *testing.T) {
	t.Parallel()
	reminder := Reminder{
		ID:           1,
		ChatID:       2,
		UserID:       3,
		Text:         "sagdshyjghj",
		RemindAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:       ReminderStatusPending,
		AttemptsLeft: 3,
	}
	assert.Equal(t, "1. 2020-01-01 03:00: sagdshyjghj", reminder.UserFormat())
}

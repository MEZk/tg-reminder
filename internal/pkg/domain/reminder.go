package domain

import (
	"fmt"
	"time"
)

// DefaultAttemptsLeft - default attempts left to deliver a reminder.
const DefaultAttemptsLeft = 3

// Reminder - reminder representation.
type Reminder struct {
	ID           int64          `db:"id"`
	ChatID       int64          `db:"chat_id"`
	UserID       int64          `db:"user_id"`
	Text         string         `db:"text"`
	CreatedAt    time.Time      `db:"created_at"`
	ModifiedAt   time.Time      `db:"modified_at"`
	RemindAt     time.Time      `db:"remind_at"`
	Status       ReminderStatus `db:"status"`
	AttemptsLeft byte           `db:"attempts_left"`
}

func (r Reminder) String() string {
	return fmt.Sprintf("[ID: %d, UserID: %d, ChatID: %d, Status: %s, RemindAt: %s, AttemptsLeft: %d, Data: %s]", r.ID, r.UserID, r.ChatID, r.Status, r.RemindAt, r.AttemptsLeft, r.Text)
}

// UserFormat - format reminder info to send to user.
func (r Reminder) UserFormat() string {
	return fmt.Sprintf("%d. %s: %s", r.ID, MoscowTime(r.RemindAt).Format(LayoutRemindAt), r.Text)
}

// ReminderStatus - status of a remidner.
type ReminderStatus string

const (
	// ReminderStatusPending is a pending status. Reminder will be sent to user at remindAt time.
	ReminderStatusPending ReminderStatus = "pending"
	// ReminderStatusDone is a done status. Reminder was sent to user and user marked it as 'done'.
	ReminderStatusDone ReminderStatus = "done"
	// ReminderStatusAttemptsExhausted describes the situation in which all attempts to receive 'done' from user are finished.
	ReminderStatusAttemptsExhausted ReminderStatus = "attempts_exhausted"
)

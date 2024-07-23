package domain

import (
	"fmt"
	"time"
)

const DefaultAttemptsLeft = 3

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

func (r Reminder) UserFormat() string {
	return fmt.Sprintf("%d. %s: %s", r.ID, MoscowTime(r.RemindAt).Format(LayoutRemindAt), r.Text)
}

type ReminderStatus string

const (
	ReminderStatusPending           ReminderStatus = "pending"
	ReminderStatusDone              ReminderStatus = "done"
	ReminderStatusAttemptsExhausted ReminderStatus = "attempts_exhausted"
)

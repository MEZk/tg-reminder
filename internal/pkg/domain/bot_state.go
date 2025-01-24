package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// BotState describes bot's current state.
type BotState struct {
	UserID     int64            `db:"user_id"`
	Name       BotStateName     `db:"name"`
	ModifiedAt time.Time        `db:"modified_at"`
	Context    *BotStateContext `db:"context"`
}

// ReminderID returns reminder id associated with current bot state.
func (s BotState) ReminderID() int64 {
	if s.Context != nil {
		return s.Context.ReminderID
	}
	return 0
}

// String implements [fmt.Stringer].
func (s BotState) String() string {
	return fmt.Sprintf("[UserID: %d, Name: %s]", s.UserID, s.Name)
}

// SetReminderID associate reminder id with current bot state..
func (s *BotState) SetReminderID(id int64) {
	if s == nil {
		return
	}

	if s.Context == nil {
		s.Context = &BotStateContext{}
	}

	s.Context.ReminderID = id
}

// SetReminderText associate reminder text with current bot state..
func (s *BotState) SetReminderText(text string) {
	if s == nil {
		return
	}

	if s.Context == nil {
		s.Context = &BotStateContext{}
	}

	s.Context.ReminderText = text
}

// ReminderText returns reminder text associated with current bot state.
func (s BotState) ReminderText() string {
	if s.Context == nil {
		return ""
	}

	return s.Context.ReminderText
}

// BotStateName is a name of bot state.
type BotStateName string

const (
	// BotStateNameStart - user sent /start command. Client is registered.
	BotStateNameStart BotStateName = "start"
	// BotStateNameHelp - user sent /help command.
	BotStateNameHelp BotStateName = "help"
	// BotStateNameCreateReminder - user sent /create_reminder command. This is an entrypoint state to create reminder.
	BotStateNameCreateReminder BotStateName = "create_reminder"
	// BotStateNameEditReminder - user clicked on reminder button .
	BotStateNameEditReminder BotStateName = "edit_reminder"
	// BotStateNameEditReminderAskAction - bot asks edti action.
	BotStateNameEditReminderAskAction BotStateName = "edit_reminder_ask_action"
	// BotStateNameRemoveReminder - user clicked on remode reminder button.
	BotStateNameRemoveReminder BotStateName = "remove_reminder"
	// BotStateNameMyReminders - user sent /my_remidners command.
	BotStateNameMyReminders BotStateName = "my_reminders"
	// BotStateNameEnableReminders - user sent /enable_reminders command.
	BotStateNameEnableReminders BotStateName = "enable_reminders"
	// BotStateNameDisableReminders - user sent /disable_reminders command.
	BotStateNameDisableReminders BotStateName = "disable_reminders"
	// BotStateNameEnterReminAt - bot is waiting on user entering remindAt.
	BotStateNameEnterReminAt BotStateName = "enter_remind_at"
)

// BotStateContext is a metadata associated with c\urrent bot state.
type BotStateContext struct {
	ReminderID   int64  `json:"reminder_id,omitempty"`
	ReminderText string `json:"reminder_text,omitempty"`
}

// Scan implements [sql.Scanner].
func (b *BotStateContext) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &b)
}

// Value implements [driver.Valuer].
func (b BotStateContext) Value() (driver.Value, error) {
	return json.Marshal(b)
}

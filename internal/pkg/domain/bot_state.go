package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type BotState struct {
	UserID     int64            `db:"user_id"`
	Name       BotStateName     `db:"name"`
	ModifiedAt time.Time        `db:"modified_at"`
	Context    *BotStateContext `db:"context"`
}

func (s BotState) ReminderID() int64 {
	if s.Context != nil {
		return s.Context.ReminderID
	}
	return 0
}

func (s BotState) String() string {
	return fmt.Sprintf("[UserID: %d, Name: %s]", s.UserID, s.Name)
}

func (s *BotState) SetReminderID(id int64) {
	if s == nil {
		return
	}

	if s.Context == nil {
		s.Context = &BotStateContext{}
	}

	s.Context.ReminderID = id
}

func (s *BotState) SetReminderText(text string) {
	if s == nil {
		return
	}

	if s.Context == nil {
		s.Context = &BotStateContext{}
	}

	s.Context.ReminderText = text
}

func (s BotState) ReminderText() string {
	if s.Context == nil {
		return ""
	}

	return s.Context.ReminderText
}

type BotStateName string

const (
	BotStateNameStart            BotStateName = "start"
	BotStateNameHelp             BotStateName = "help"
	BotStateNameCreateReminder   BotStateName = "create_reminder"
	BotStateNameEditReminder     BotStateName = "edit_reminder"
	BotStateNameRemoveReminder   BotStateName = "remove_reminder"
	BotStateNameMyReminders      BotStateName = "my_reminders"
	BotStateNameEnableReminders  BotStateName = "enable_reminders"
	BotStateNameDisableReminders BotStateName = "disable_reminders"

	BotStateNameEnterReminAt BotStateName = "enter_remind_at"
)

type BotStateContext struct {
	ReminderID   int64  `json:"reminder_id,omitempty"`
	ReminderText string `json:"reminder_text,omitempty"`
}

func (b *BotStateContext) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &b)
}

func (b BotStateContext) Value() (driver.Value, error) {
	return json.Marshal(b)
}

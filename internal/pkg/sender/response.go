package sender

import (
	"fmt"
	"strings"
)

// BotResponse describes bot's reaction on particular message in chat.
type BotResponse struct {
	ChatID           int64  // telegram chat id
	ReplyToMessageID int64  // message to reply to, if 0 then no reply but common message
	Text             string // message text

	showMyReminderListEditButtons bool
	showReminderDatesButtons      bool
	showReminderDoneButtons       bool
	reminderID                    int64
}

// BotResponseOption - describes response option.
type BotResponseOption func(r *BotResponse)

// WithMyRemindersListEditButtons - shows inline keyboard with my reminders list edit buttons.
// "Edit" button to edit reminder and "Remove" button to remove reminder.
func WithMyRemindersListEditButtons() BotResponseOption {
	return func(r *BotResponse) {
		r.showMyReminderListEditButtons = true
	}
}

// WithReminderDatesButtons - shows inline keyboard to set date to fire reminder.
func WithReminderDatesButtons() BotResponseOption {
	return func(r *BotResponse) {
		r.showReminderDatesButtons = true
	}
}

// WithReminderDoneButton - shows inline keyboard to allow user to mark reminder with specific id as done.
func WithReminderDoneButton(reminderID int64) BotResponseOption {
	return func(r *BotResponse) {
		r.showReminderDoneButtons = true
		r.reminderID = reminderID
	}
}

func (s BotResponse) String() string {
	text := strings.ReplaceAll(s.Text, "\n", "\\n")

	if s.reminderID == 0 {
		return fmt.Sprintf("[ChatID: %d, ReplyToMessageID: %d, Text: %s]", s.ChatID, s.ReplyToMessageID, text)
	}

	return fmt.Sprintf("[ChatID: %d, ReplyToMessageID: %d, RemidnerID: %d, Text: %s]", s.ChatID, s.ReplyToMessageID, s.reminderID, text)
}

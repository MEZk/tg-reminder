package domain

import (
	"fmt"
	"strings"
	"time"
)

// TgMessage represents a Telegram message.
// See [github.com/go-telegram-bot-api/telegram-bot-api/v5.Message].
type TgMessage struct {
	ChatID   int64
	UserID   int64
	UserName string
	Text     string
}

// IsCommand returns true if message is a command (starts with "/").
func (m TgMessage) IsCommand() bool {
	return strings.HasPrefix(m.Text, "/")
}

// String implements [fmt.Stringer].
func (m TgMessage) String() string {
	return fmt.Sprintf("[ChatID: %d, UserID: %d, UserName: %s, Text: %s]", m.ChatID, m.UserID, m.UserName, m.Text)
}

// RemindAt extracts date and time when reminder should be sent to user.
func (m TgMessage) RemindAt() (time.Time, error) {
	// TODO: parse remindAt in user timezone, now default is MSK
	remindAt, err := time.ParseInLocation(LayoutRemindAt, m.Text, locationMSK)
	if err != nil {
		return time.Time{}, fmt.Errorf("can't parse remindAt %s: %w", m.Text, err)
	}

	return remindAt.Truncate(1 * time.Minute), nil
}

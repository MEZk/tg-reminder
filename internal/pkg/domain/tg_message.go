package domain

import (
	"fmt"
	"strings"
	"time"
)

type TgMessage struct {
	ChatID   int64
	UserID   int64
	UserName string
	Text     string
}

func (m TgMessage) IsCommand() bool {
	return strings.HasPrefix(m.Text, "/")
}

func (m TgMessage) String() string {
	return fmt.Sprintf("[ChatID: %d, UserID: %d, UserName: %s, Text: %s]", m.ChatID, m.UserID, m.UserName, m.Text)
}

func (m TgMessage) RemindAt() (time.Time, error) {
	// TODO: parse remindAt in user timezone, now default is MSK
	remindAt, err := time.ParseInLocation(LayoutRemindAt, m.Text, locationMSK)
	if err != nil {
		return time.Time{}, fmt.Errorf("can't parse remindAt %s: %w", m.Text, err)
	}

	return remindAt.Truncate(1 * time.Minute), nil
}

package domain

import (
	"fmt"
	"strings"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
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
func (m TgMessage) RemindAt(now time.Time) (time.Time, error) {
	// TODO: parse remindAt in user timezone, now default is MSK
	now = MoscowTime(now)

	remindAt, err := time.ParseInLocation(LayoutRemindAt, m.Text, now.Location())
	if err != nil {
		log.Printf("[DEBUG] can't parse (time.ParseInLocation) remindAt %s: %s\n", m.Text, err)

		var remindAtDate date.Date
		if remindAtDate, err = dateparser.Parse(&dateparser.Configuration{
			CurrentTime:         now,
			Locales:             []string{"ru"},
			PreferredDateSource: dateparser.Future,
		}, m.Text); err != nil {
			return time.Time{}, fmt.Errorf("can't parse (go-dateparser) remindAt: %w", err)
		}

		return remindAtDate.Time, nil
	}

	return remindAt.Truncate(1 * time.Minute), nil
}

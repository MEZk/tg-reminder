package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TgCallbackQuery represents an incoming callback query from a callback button in
// an inline keyboard. See [github.com/go-telegram-bot-api/telegram-bot-api/v5.CallbackQuery].
type TgCallbackQuery struct {
	ChatID   int64
	UserID   int64
	UserName string
	Data     string
}

const (
	// ButtonDataPrefixRemindAtTime - button prefix for [domain.TgCallbackQuery] data which contains remindAt formatted as time.
	ButtonDataPrefixRemindAtTime = "btn_remind_at/time/"
	// ButtonDataPrefixRemindAtDuration - button prefix for [domain.TgCallbackQuery] data which contains remindAt formatted as duration.
	ButtonDataPrefixRemindAtDuration = "btn_remind_at/duration/"
	// ButtonDataPrefixReminderDone - button prefix for [domain.TgCallbackQuery] data which contains id of reminder to mark it as done.
	ButtonDataPrefixReminderDone = "btn_reminder_done/"
	// ButtonDataPrefixDelayReminder - button prefix for [domain.TgCallbackQuery] data which contains duration to delay reminder.
	ButtonDataPrefixDelayReminder = "btn_delay_reminder/"

	// ButtonDataEditReminder - [domain.TgCallbackQuery] data for edit reminder button.
	ButtonDataEditReminder = "btn_edit_reminder"
	// ButtonDataRemoveReminder - [domain.TgCallbackQuery] data for remove reminder button.
	ButtonDataRemoveReminder = "btn_remove_reminder"
)

// IsButtonClick returns true, if callback query is a known button click.
func (q TgCallbackQuery) IsButtonClick() bool {
	return strings.HasPrefix(q.Data, "btn_")
}

// String implements [fmt.Stringer].
func (q TgCallbackQuery) String() string {
	return fmt.Sprintf("[ChatID: %d, UserID: %d, UserName: %s, Data: %s]", q.ChatID, q.UserID, q.UserName, q.Data)
}

// LayoutRemindAt is the layout to format and parse [domain.Reminder] RemindAt.
const LayoutRemindAt = "2006-01-02 15:04"

// RemindAt extracts date and time when reminder should be sent to user.
func (q TgCallbackQuery) RemindAt(now time.Time) (time.Time, error) {
	now = MoscowTime(now)

	if timeSuffix, ok := strings.CutPrefix(q.Data, ButtonDataPrefixRemindAtTime); ok {
		remindAtTime, err := time.Parse("15:04", timeSuffix)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse remindAt time: %w", err)
		}

		remindAt := time.Date(now.Year(), now.Month(), now.Day(), remindAtTime.Hour(), remindAtTime.Minute(), now.Second(), now.Nanosecond(), now.Location())

		// if remindAt is before now we suppose that it should be tomorrow
		if remindAt.Before(now) {
			remindAt = remindAt.Add(24 * time.Hour)
		}

		return remindAt.Truncate(1 * time.Minute), nil
	}

	if durationSuffix, ok := strings.CutPrefix(q.Data, ButtonDataPrefixRemindAtDuration); ok {
		remindAtDuration, err := time.ParseDuration(durationSuffix)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse remindAt duration: %w", err)
		}

		return now.Add(remindAtDuration).Truncate(1 * time.Minute), nil
	}

	if suffix, ok := strings.CutPrefix(q.Data, ButtonDataPrefixDelayReminder); ok {
		if fields := strings.Split(suffix, "/"); len(fields) == 2 {
			remindAtDuration, err := time.ParseDuration(fields[1])
			if err != nil {
				return time.Time{}, fmt.Errorf("failed to parse remindAt duration: %w", err)
			}

			return now.Add(remindAtDuration).Truncate(1 * time.Minute), nil
		}
	}

	return time.Time{}, fmt.Errorf("unknown remindAt format: %s", q.Data)
}

// ReminderID extracts reminder id.
func (q TgCallbackQuery) ReminderID() (int64, error) {
	if idSuffix, ok := strings.CutPrefix(q.Data, ButtonDataPrefixReminderDone); ok {
		return strconv.ParseInt(idSuffix, 10, 64)
	}

	if suffix, ok := strings.CutPrefix(q.Data, ButtonDataPrefixDelayReminder); ok {
		if fields := strings.Split(suffix, "/"); len(fields) == 2 {
			return strconv.ParseInt(fields[0], 10, 64)
		}
	}

	return 0, fmt.Errorf("unknown reminder id format: %s", q.Data)
}

// IsRemindAtButtonClick returns true is callback is a button click to set remindAt.
func (q TgCallbackQuery) IsRemindAtButtonClick() bool {
	return strings.HasPrefix(q.Data, ButtonDataPrefixRemindAtTime) ||
		strings.HasPrefix(q.Data, ButtonDataPrefixRemindAtDuration)
}

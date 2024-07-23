package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TgCallbackQuery struct {
	ChatID   int64
	UserID   int64
	UserName string
	Data     string
}

const (
	ButtonDataPrefixRemindAtTime     = "btn_remind_at/time/"
	ButtonDataPrefixRemindAtDuration = "btn_remind_at/duration/"
	ButtonDataPrefixReminderDone     = "btn_reminder_done/"
	ButtonDataPrefixDelayReminder    = "btn_delay_reminder/"

	ButtonDataEditReminder   = "btn_edit_reminder"
	ButtonDataRemoveReminder = "btn_remove_reminder"
)

func (q TgCallbackQuery) IsButtonClick() bool {
	return strings.HasPrefix(q.Data, "btn_")
}

func (q TgCallbackQuery) String() string {
	return fmt.Sprintf("[ChatID: %d, UserID: %d, UserName: %s, Data: %s]", q.ChatID, q.UserID, q.UserName, q.Data)
}

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
}

const LayoutRemindAt = "2006-01-02 15:04"

func (q TgCallbackQuery) RemindAt(now time.Time) (time.Time, error) {
	now = MoscowTime(now)

	if timeSuffix, ok := strings.CutPrefix(q.Data, ButtonDataPrefixRemindAtTime); ok {
		remindAtTime, err := time.Parse("15:04", timeSuffix)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse remindAt time: %w", err)
		}

		remindAt := time.Date(now.Year(), now.Month(), now.Day(), remindAtTime.Hour(), remindAtTime.Minute(), now.Second(), now.Nanosecond(), now.Location())

		// if remindAt is before now we suppose that it should be tommorow
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

func (q TgCallbackQuery) IsRemindAtButtonClick() bool {
	return strings.HasPrefix(q.Data, ButtonDataPrefixRemindAtTime) ||
		strings.HasPrefix(q.Data, ButtonDataPrefixRemindAtDuration)
}

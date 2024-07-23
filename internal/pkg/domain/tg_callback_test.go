package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTgCallbackQuery_IsButtonClick(t *testing.T) {
	t.Parallel()
	assert.True(t, TgCallbackQuery{Data: "btn_"}.IsButtonClick())
	assert.False(t, TgCallbackQuery{Data: "/start"}.IsButtonClick())
}

func TestTgCallbackQuery_IsRemindAtButtonClick(t *testing.T) {
	t.Parallel()
	assert.True(t, TgCallbackQuery{Data: "btn_remind_at/time/"}.IsRemindAtButtonClick())
	assert.True(t, TgCallbackQuery{Data: "btn_remind_at/duration/"}.IsRemindAtButtonClick())
	assert.False(t, TgCallbackQuery{Data: "btn_reminder_done"}.IsRemindAtButtonClick())
}

func TestTgCallbackQuery_RemindAt(t *testing.T) {
	testCases := []struct {
		name   string
		now    time.Time
		query  TgCallbackQuery
		expRes time.Time
		expErr string
	}{
		{
			name:   "success: remind at time",
			now:    time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC),
			query:  TgCallbackQuery{Data: "btn_remind_at/time/11:30"},
			expRes: time.Date(2024, 1, 1, 11, 30, 0, 0, locationMSK),
		},
		{
			name:   "success: remind at time is in past",
			now:    time.Date(2024, 1, 1, 12, 0, 1, 0, time.UTC),
			query:  TgCallbackQuery{Data: "btn_remind_at/time/11:30"},
			expRes: time.Date(2024, 1, 2, 11, 30, 0, 0, locationMSK),
		},
		{
			name:   "success: remind at duration",
			now:    time.Date(2024, 1, 1, 12, 0, 1, 0, time.UTC),
			query:  TgCallbackQuery{Data: "btn_remind_at/duration/1h"},
			expRes: time.Date(2024, 1, 1, 16, 0, 0, 0, locationMSK),
		},
		{
			name:   "success: delay reminder",
			now:    time.Date(2024, 1, 1, 12, 0, 1, 0, time.UTC),
			query:  TgCallbackQuery{Data: "btn_delay_reminder/1234/1h"},
			expRes: time.Date(2024, 1, 1, 16, 0, 0, 0, locationMSK),
		},
		{
			name:   "error: remind at time, can't parse",
			query:  TgCallbackQuery{Data: "btn_remind_at/time/aaaaa"},
			expErr: `failed to parse remindAt time: parsing time "aaaaa" as "15:04": cannot parse "aaaaa" as "15"`,
		},
		{
			name:   "error: remind at duration, can't parse",
			query:  TgCallbackQuery{Data: "btn_remind_at/duration/aaaaa"},
			expErr: `failed to parse remindAt duration: time: invalid duration "aaaaa"`,
		},
		{
			name:   "error: delay reminder, can't parse",
			query:  TgCallbackQuery{Data: "btn_delay_reminder/1234/121r4gfsg"},
			expErr: `failed to parse remindAt duration: time: unknown unit "r" in duration "121r4gfsg"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actRes, actErr := tc.query.RemindAt(tc.now)
			assert.Equal(t, tc.expRes, actRes)
			if tc.expErr != "" {
				assert.EqualError(t, actErr, tc.expErr)
			} else {
				assert.NoError(t, actErr)
			}
		})
	}
}

func TestTgCallbackQuery_ReminderID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		query  TgCallbackQuery
		expRes int64
		expErr string
	}{
		{
			name: "Done button click",
			query: TgCallbackQuery{
				Data: "btn_reminder_done/1234",
			},
			expRes: 1234,
		},
		{
			name: "Delay button click",
			query: TgCallbackQuery{
				Data: "btn_delay_reminder/1234/11:30",
			},
			expRes: 1234,
		},
		{
			name: "unknown format",
			query: TgCallbackQuery{
				Data: "foo",
			},
			expErr: "unknown reminder id format: foo",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actRes, actErr := tc.query.ReminderID()
			assert.Equal(t, tc.expRes, actRes)
			if tc.expErr != "" {
				assert.EqualError(t, actErr, tc.expErr)
			} else {
				assert.NoError(t, actErr)
			}
		})
	}
}

func TestTgCallbackQuery_String(t *testing.T) {
	t.Parallel()
	query := TgCallbackQuery{
		ChatID:   1,
		UserID:   2,
		UserName: "John Doe",
		Data:     "data",
	}
	assert.Equal(t, "[ChatID: 1, UserID: 2, UserName: John Doe, Data: data]", query.String())
}

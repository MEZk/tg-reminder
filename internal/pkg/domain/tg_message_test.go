package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTgMessage_IsCommand(t *testing.T) {
	t.Parallel()
	assert.True(t, TgMessage{Text: "/foo"}.IsCommand())
	assert.False(t, TgMessage{Text: "bar"}.IsCommand())
}

func TestTgMessage_RemindAt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		now    time.Time
		msg    TgMessage
		expRes time.Time
		expErr string
	}{
		{
			name:   "success: absolute full date",
			now:    time.Date(2024, 8, 18, 0, 0, 0, 0, time.UTC),
			msg:    TgMessage{Text: "2024-08-18 11:00"},
			expRes: time.Date(2024, 8, 18, 11, 0, 0, 0, locationMSK),
		},
		{
			name:   "success: relative Russian date (19:00)",
			now:    time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg:    TgMessage{Text: "19:00"},
			expRes: time.Date(2024, 8, 18, 19, 0, 0, 0, locationMSK),
		},
		{
			name:   "success: relative Russian date (tomorrow)",
			now:    time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg:    TgMessage{Text: "завтра"},
			expRes: time.Date(2024, 8, 19, 0, 0, 0, 0, locationMSK),
		},
		{
			name:   "success: relative Russian date (tomorrow at 19:00)",
			now:    time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg:    TgMessage{Text: "завтра в 15:00"},
			expRes: time.Date(2024, 8, 19, 15, 0, 0, 0, locationMSK),
		},
		{
			name:   "success: relative Russian date (Wed at 15:00)",
			now:    time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg:    TgMessage{Text: "в среду в 15:00"},
			expRes: time.Date(2024, 8, 21, 15, 0, 0, 0, locationMSK),
		},
		{
			name: "success: relative russian date (in 1 hour)",
			now:  time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg: TgMessage{
				Text: "через час",
			},
			expRes: time.Date(2024, 8, 18, 1, 0, 0, 0, locationMSK),
		},
		{
			name: "success: relative russian date (in 2 hours)",
			now:  time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg: TgMessage{
				Text: "через 2 часа",
			},
			expRes: time.Date(2024, 8, 18, 2, 0, 0, 0, locationMSK),
		},
		{
			name:   "success: 30.01.2024 at 11:00",
			now:    time.Date(2024, 8, 18, 0, 0, 0, 0, time.UTC),
			msg:    TgMessage{Text: "30.01.2024 в 11:00"},
			expRes: time.Date(2024, 1, 30, 11, 0, 0, 0, locationMSK),
		},
		{
			name:   "success: relative Russian date (in 1 month)",
			now:    time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg:    TgMessage{Text: "через месяц"},
			expRes: time.Date(2024, 9, 18, 0, 0, 0, 0, locationMSK),
		},
		{
			name: "success: relative Russian date (in 1 week)",
			now:  time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg: TgMessage{
				Text: "через неделю",
			},
			expRes: time.Date(2024, 8, 25, 0, 0, 0, 0, locationMSK),
		},
		{
			name: "success: relative Russian date (in 1 week)",
			now:  time.Date(2024, 8, 18, 0, 0, 0, 0, locationMSK),
			msg: TgMessage{
				Text: "в 20:00",
			},
			expRes: time.Date(2024, 8, 18, 20, 0, 0, 0, locationMSK),
		},
		{
			name: "error: can't parse",
			now:  time.Date(2024, 8, 18, 0, 0, 0, 0, time.UTC),
			msg: TgMessage{
				Text: "foo bar baz",
			},
			expErr: `can't parse (go-dateparser) remindAt: failed to parse "foo bar baz": unknown format`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)

			actRes, actErr := tc.msg.RemindAt(tc.now)

			if tc.expErr == "" {
				r.NoError(actErr)
			} else {
				r.EqualError(actErr, tc.expErr)
			}

			r.Equal(tc.expRes, actRes)
		})
	}
}

func TestTgMessage_String(t *testing.T) {
	t.Parallel()
	msg := TgMessage{
		ChatID:   25,
		UserID:   213,
		UserName: "John Doe",
		Text:     "Foo Bar",
	}
	assert.Equal(t, "[ChatID: 25, UserID: 213, UserName: John Doe, Text: Foo Bar]", msg.String())
}

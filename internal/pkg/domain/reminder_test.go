package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DefaultAttemptsLeft(t *testing.T) {
	t.Parallel()
	require.EqualValues(t, 3, DefaultAttemptsLeft)
}

func TestReminder_String(t *testing.T) {
	t.Parallel()
	reminder := Reminder{
		ID:           1,
		ChatID:       2,
		UserID:       3,
		Text:         "sagdshyjghj",
		RemindAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:       ReminderStatusPending,
		AttemptsLeft: 3,
	}
	assert.Equal(t, "[ID: 1, UserID: 3, ChatID: 2, Status: pending, RemindAt: 2020-01-01 00:00:00 +0000 UTC, AttemptsLeft: 3, Data: sagdshyjghj]", reminder.String())

	assert.EqualValues(t, "pending", ReminderStatusPending)
	assert.EqualValues(t, "attempts_exhausted", ReminderStatusAttemptsExhausted)
	assert.EqualValues(t, "done", ReminderStatusDone)
}

func Test_getRussianMonth(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	a.Equal("янв.", getRussianMonth(1))
	a.Equal("фев.", getRussianMonth(2))
	a.Equal("мар.", getRussianMonth(3))
	a.Equal("апр.", getRussianMonth(4))
	a.Equal("мая", getRussianMonth(5))
	a.Equal("июн.", getRussianMonth(6))
	a.Equal("июл.", getRussianMonth(7))
	a.Equal("авг.", getRussianMonth(8))
	a.Equal("сент.", getRussianMonth(9))
	a.Equal("окт.", getRussianMonth(10))
	a.Equal("нояб.", getRussianMonth(11))
	a.Equal("дек.", getRussianMonth(12))
}

func TestReminder_FormatNotify(t *testing.T) {
	t.Parallel()
	r := Reminder{Text: "do some thing"}
	require.Equal(t, "‼️ *НАПОМИНАНИЕ* ‼️\n\n *DO SOME THING*\n\n⏰ Сегодня 02:30", r.FormatNotify())
}

func TestReminder_FormatList(t *testing.T) {
	t.Parallel()

	var (
		jan1 = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		jan2 = jan1.Add(24 * time.Hour)
	)

	testCases := []struct {
		name     string
		now      time.Time
		reminder Reminder
		expRes   string
	}{
		{
			name: "today",
			now:  jan1,
			reminder: Reminder{
				ID:       1,
				Text:     "Foo bar baz",
				RemindAt: jan1,
			},
			expRes: "✅ *Foo bar baz*❗\n⏰ Сегодня 03:00\n#️⃣ 1",
		},
		{
			name: "tommorow",
			now:  jan1,
			reminder: Reminder{
				ID:       1,
				Text:     "Foo bar baz",
				RemindAt: jan2,
			},
			expRes: "✅ *Foo bar baz*\n⏰ 2 янв. 03:00\n#️⃣ 1",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expRes, tc.reminder.FormatList(tc.now))
		})
	}
}

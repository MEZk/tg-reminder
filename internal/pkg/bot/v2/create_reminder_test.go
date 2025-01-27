package v2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_parseRemidnerDate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		data        string
		isUserInput bool
		now         time.Time
		exp         time.Time
		expErr      string
	}{
		{
			name: "time before now",
			data: "11:30",
			now:  time.Date(2025, 1, 1, 14, 30, 0, 0, time.UTC),
			exp:  time.Date(2025, 1, 2, 11, 30, 0, 0, time.UTC),
		},
		{
			name: "time after now",
			data: "15:30",
			now:  time.Date(2025, 1, 1, 9, 30, 0, 0, time.UTC),
			exp:  time.Date(2025, 1, 1, 15, 30, 0, 0, time.UTC),
		},
		{
			name: "duration 30m",
			data: "30m",
			now:  time.Date(2025, 1, 1, 9, 30, 0, 0, time.UTC),
			exp:  time.Date(2025, 1, 1, 10, 00, 0, 0, time.UTC),
		},
		{
			name: "duration 30m, next day",
			data: "30m",
			now:  time.Date(2025, 1, 1, 23, 35, 0, 0, time.UTC),
			exp:  time.Date(2025, 1, 2, 0, 05, 0, 0, time.UTC),
		},
		{
			name:        "user input: full date (v1)",
			data:        "2025-03-02 11:30",
			isUserInput: true,
			now:         time.Date(2025, 1, 1, 9, 30, 0, 0, time.UTC),
			exp:         time.Date(2025, 3, 2, 11, 30, 0, 0, time.UTC),
		},
		{
			name:        "user input: full date (v2)",
			data:        "02.05.2025 19:00",
			isUserInput: true,
			now:         time.Date(2025, 1, 1, 9, 30, 0, 0, time.UTC),
			exp:         time.Date(2025, 5, 2, 19, 00, 0, 0, time.UTC),
		},
		{
			name:        "user input: ru tommowor",
			data:        "Завтра в 11:00",
			isUserInput: true,
			now:         time.Date(2025, 1, 1, 9, 30, 0, 0, time.UTC),
			exp:         time.Date(2025, 1, 2, 11, 00, 0, 0, time.UTC),
		},
		{
			name:   "error: incorrect duration format",
			data:   "m30",
			now:    time.Date(2025, 1, 1, 9, 30, 0, 0, time.UTC),
			expErr: `time: invalid duration "m30"`,
		},
		{
			name:   "error: incorrect time format",
			data:   "11:30fa",
			now:    time.Date(2025, 1, 1, 9, 30, 0, 0, time.UTC),
			expErr: `failed to parse "11:30fa": unknown format`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.data, func(t *testing.T) {
			t.Parallel()
			actDate, actErr := parseRemidnerDate(tc.data, tc.now, tc.isUserInput)
			r := require.New(t)
			if tc.expErr == "" {
				r.NoError(actErr)
			} else {
				r.EqualError(actErr, tc.expErr)
			}
			r.Equal(tc.exp, actDate)
		})
	}
}

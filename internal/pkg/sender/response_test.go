package sender

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBotResponse_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		resp BotResponse
		want string
	}{
		{
			name: "reminderID = 0",
			resp: BotResponse{
				ChatID:           1,
				ReplyToMessageID: 2,
				Text:             "FooBar",
			},
			want: "[ChatID: 1, ReplyToMessageID: 2, Text: FooBar]",
		},
		{
			name: "reminderID != 0",
			resp: BotResponse{
				ChatID:           2,
				ReplyToMessageID: 4,
				Text:             "FooBarBaz",
				reminderID:       6453,
			},
			want: "[ChatID: 2, ReplyToMessageID: 4, RemidnerID: 6453, Text: FooBarBaz]",
		},
		{
			name: "text with new line chars",
			resp: BotResponse{
				ChatID:           7,
				ReplyToMessageID: 2,
				Text:             "FooBar\n",
			},
			want: `[ChatID: 7, ReplyToMessageID: 2, Text: FooBar\n]`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.New(t).Equal(tc.want, tc.resp.String())
		})
	}
}

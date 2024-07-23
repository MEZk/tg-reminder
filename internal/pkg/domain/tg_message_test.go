package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTgMessage_IsCommand(t *testing.T) {
	t.Parallel()
	assert.True(t, TgMessage{Text: "/foo"}.IsCommand())
	assert.False(t, TgMessage{Text: "bar"}.IsCommand())
}

func TestTgMessage_RemindAt(t *testing.T) {
	t.Parallel()

	remindAt, err := TgMessage{Text: "2024-06-13 11:22"}.RemindAt()
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2024, 6, 13, 11, 22, 0, 0, locationMSK), remindAt)

	remindAt, err = TgMessage{Text: "foo"}.RemindAt()
	assert.EqualError(t, err, `can't parse remindAt foo: parsing time "foo" as "2006-01-02 15:04": cannot parse "foo" as "2006"`)

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

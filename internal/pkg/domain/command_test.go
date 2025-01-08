package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBotCommand_Markdown(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "/create\\_reminder", BotCommandCreateReminder.Markdown())
	assert.Equal(t, "/start", BotCommandStart.Markdown())
}

func TestBotCommand_String(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	a.Equal("/start", BotCommandStart.String())
	a.Equal("/help", BotCommandHelp.String())
	a.Equal("/create_reminder", BotCommandCreateReminder.String())
	a.Equal("/my_reminders", BotCommandMyReminders.String())
	a.Equal("/enable_reminders", BotCommandEnableReminders.String())
	a.Equal("/disable_reminders", BotCommandDisableReminders.String())
}

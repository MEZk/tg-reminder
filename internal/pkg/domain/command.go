package domain

import (
	"strings"
)

type BotCommand string

const (
	BotCommandStart            BotCommand = "/start"
	BotCommandHelp             BotCommand = "/help"
	BotCommandCreateReminder   BotCommand = "/create_reminder"
	BotCommandMyReminders      BotCommand = "/my_reminders"
	BotCommandEnableReminders  BotCommand = "/enable_reminders"
	BotCommandDisableReminders BotCommand = "/disable_reminders"
)

func (c BotCommand) String() string {
	return string(c)
}

func (c BotCommand) Markdown() string {
	return strings.Replace(string(c), "_", "\\_", -1)
}

package domain

import (
	"strings"
)

// BotCommand is a command which can be processed by bot.
type BotCommand string

const (
	// BotCommandStart is a start command.
	BotCommandStart BotCommand = "/start"
	// BotCommandHelp is a help command.
	BotCommandHelp BotCommand = "/help"
	// BotCommandCreateReminder is a command to start reminder creation process.
	BotCommandCreateReminder BotCommand = "/create_reminder"
	// BotCommandMyReminders is a command to show all [domain.ReminderStatusPending] reminders for user.
	BotCommandMyReminders BotCommand = "/my_reminders"
	// BotCommandEnableReminders is a command to disable all reminders for user. User status will be chanhed to [domain.UserStatusInactive].
	BotCommandEnableReminders BotCommand = "/enable_reminders"
	// BotCommandDisableReminders - is a command to enable all reminders for user. User status will be chanhed to [domain.UserStatusActive].
	BotCommandDisableReminders BotCommand = "/disable_reminders"
)

// String implememts [fmt.Stringer].
func (c BotCommand) String() string {
	return string(c)
}

// Markdown formats bot command as a Markdown.
func (c BotCommand) Markdown() string {
	return strings.Replace(string(c), "_", "\\_", -1)
}

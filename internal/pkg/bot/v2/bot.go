package v2

import (
	"context"
	"fmt"

	log "github.com/go-pkgz/lgr"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

// Storage – storage interface used by bot.
type Storage interface {
	SaveUser(ctx context.Context, user domain.User) error
	SetUserStatus(ctx context.Context, id int64, status domain.UserStatus) error
	SaveReminder(ctx context.Context, reminder domain.Reminder) (int64, error)
	GetMyReminders(ctx context.Context, userID, chatID int64) ([]domain.Reminder, error)
	RemoveReminder(ctx context.Context, id int64) error
}

type tgBot struct {
	store Storage
	api   *bot.Bot
	state *fsm.FSM
}

// New creates and initialize bot.
// Initialization means send getMe (see https://core.telegram.org/bots/api#getme)
// and setMyCommands (see https://core.telegram.org/bots/api#setmycommands) requests to Telegram API.
// Returns error if request to Telegram API failed.
// Set debug flag to true if you need request/response logs.
func New(ctx context.Context, store Storage, token string, debug bool) (*tgBot, error) {
	tBot := tgBot{
		store: store,
		state: fsm.New(stateDefault, map[fsm.StateID]fsm.Callback{}),
	}

	opts := []bot.Option{bot.WithDefaultHandler(tBot.defaultHandler)}
	if debug {
		opts = append(opts, bot.WithDebug(), bot.WithDebugHandler(tgAPIDebugLogHandler))
	}

	api, err := bot.New(token, opts...)
	if err != nil {
		return nil, err
	}

	tBot.api = api

	if _, err = api.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: commandStart, Description: "Начать работу с ботом"},
			{Command: commandHelp, Description: "Справка"},
			{Command: commandCreateReminder, Description: "Создать напоминание"},
			{Command: commandMyReminders, Description: "Мои напоминания"},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to set bot commands: %w", err)
	}

	tBot.api.RegisterHandler(bot.HandlerTypeMessageText, commandStart, bot.MatchTypeExact, tBot.startHandler)
	tBot.api.RegisterHandler(bot.HandlerTypeMessageText, commandHelp, bot.MatchTypeExact, tBot.helpHandler)

	tBot.api.RegisterHandler(bot.HandlerTypeMessageText, commandCreateReminder, bot.MatchTypeExact, tBot.createReminderHandler)
	tBot.api.RegisterHandler(bot.HandlerTypeMessageText, commandMyReminders, bot.MatchTypeExact, tBot.myRemindersHandler)

	return &tBot, nil
}

// Start starts receiving update from Telegram API.
func (tb *tgBot) Start(ctx context.Context) {
	tb.api.Start(ctx)
	log.Print("[DEBUG] bot started")
}

func (tb *tgBot) sendMessage(ctx context.Context, msg *bot.SendMessageParams) {
	if msg.ParseMode == "" {
		msg.ParseMode = models.ParseModeHTML
	}

	if _, err := tb.api.SendMessage(ctx, msg); err != nil {
		log.Printf("[WARN] failed to send message for user %d: %v", msg.ChatID, err)
	}
}

func tgAPIDebugLogHandler(format string, args ...any) {
	log.Printf("[DEBUG]: "+format, args...)
}

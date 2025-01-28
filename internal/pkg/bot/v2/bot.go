package v2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

const (
	cmdStart          = "/start"
	cmdHelp           = "/help"
	cmdCreateReminder = "/create_reminder"
	cmdMyReminders    = "/my_reminders"
)

type Storage interface {
	GetBotState(ctx context.Context, userID int64) (domain.BotState, error)
	SaveBotState(ctx context.Context, state domain.BotState) error

	SaveUser(ctx context.Context, user domain.User) error

	SaveReminder(ctx context.Context, reminder domain.Reminder) (int64, error)
	GetMyReminders(ctx context.Context, userID, chatID int64) ([]domain.Reminder, error)
}

type tgBot struct {
	store Storage
	stop  func()
	start func()
}

func New(store Storage, token string) (*tgBot, error) {
	pref := tele.Settings{
		Token:     token,
		ParseMode: tele.ModeHTML,
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	b.Use(middleware.Recover(), middleware.AutoRespond())

	tb := &tgBot{
		store: store,
		stop:  b.Stop,
		start: b.Start,
	}

	if err = b.SetCommands([]tele.Command{
		{Text: cmdStart, Description: "Начать работу"},
		{Text: cmdHelp, Description: "Справка"},
		{Text: cmdCreateReminder, Description: "Создать напоминания"},
		{Text: cmdMyReminders, Description: "Мои напоминания"},
	}); err != nil {
		return nil, fmt.Errorf("failed to set bot commands: %w", err)
	}

	tb.registerHandlers(b)

	return tb, nil
}

func (b *tgBot) Start() {
	log.Printf("[INFO] bot started]")
	b.start()
}

func (b *tgBot) Stop() {
	b.stop()
	log.Print("[INFO] bot stopped")
}

func (tb *tgBot) registerHandlers(b *tele.Bot) {
	b.Handle(cmdStart, tb.onStartCmd)
	b.Handle(cmdHelp, tb.onHelpCmd)
	b.Handle(cmdCreateReminder, tb.onCreateReminderCmd)
	b.Handle(cmdMyReminders, tb.onMyRemindersCmd)
	b.Handle(tele.OnText, tb.onUserMessage)

	// Reminder date inline keyboard
	b.Handle(&btn11_30, tb.onReminderDateBtn)
	b.Handle(&btn14_30, tb.onReminderDateBtn)
	b.Handle(&btn19_30, tb.onReminderDateBtn)
	b.Handle(&btn20_30, tb.onReminderDateBtn)
	b.Handle(&btn30m, tb.onReminderDateBtn)
	b.Handle(&btn80m, tb.onReminderDateBtn)
	b.Handle(&btn1day, tb.onReminderDateBtn)
	b.Handle(&btn1mon, tb.onReminderDateBtn)
}

func (b *tgBot) onUserMessage(c tele.Context) error {
	userID := c.Sender().ID
	ctx := context.TODO()

	state, err := b.store.GetBotState(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get bot state for user %d: %w", userID, err)
	}

	switch state.Name {
	case domain.BotStateNameCreateReminder:
		return b.onReminderTextReceived(ctx, c, state)
	case domain.BotStateNameEnterReminAt:
		return b.onRemindDateReceived(ctx, c, state)
	default:
		return fmt.Errorf("unknown bot state '%s' for user %d", state, userID)
	}
}

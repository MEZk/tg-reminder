package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
)

// ResponseSender - bot's response sender.
type ResponseSender interface {
	SendBotResponse(response sender.BotResponse, opts ...sender.BotResponseOption) error
}

type Storage interface {
	GetBotState(ctx context.Context, userID int64) (domain.BotState, error)
	SaveBotState(ctx context.Context, state domain.BotState) error

	SaveUser(ctx context.Context, user domain.User) error
	SetUserStatus(ctx context.Context, id int64, inactive domain.UserStatus) error

	SaveReminder(ctx context.Context, reminder domain.Reminder) (int64, error)
	GetMyReminders(ctx context.Context, userID, chatID int64) ([]domain.Reminder, error)
	RemoveReminder(ctx context.Context, id int64) error
	SetReminderStatus(ctx context.Context, id int64, status domain.ReminderStatus) error
	DelayReminder(ctx context.Context, id int64, remindAt time.Time) error
}

type Bot struct {
	responseSender ResponseSender
	store          Storage
}

func New(responseSender ResponseSender, store Storage) *Bot {
	return &Bot{responseSender: responseSender, store: store}
}

func (b *Bot) OnMessage(ctx context.Context, message domain.TgMessage) error {
	if message.IsCommand() {
		switch message.Text {
		case domain.BotCommandStart.String():
			return b.onStartCommand(ctx, message)
		case domain.BotCommandHelp.String():
			return b.onHelpCommand(ctx, message)
		case domain.BotCommandCreateReminder.String():
			return b.onCreateReminderCommand(ctx, message)
		case domain.BotCommandMyReminders.String():
			return b.onMyRemindersCommand(ctx, message)
		case domain.BotCommandEnableReminders.String():
			return b.onEnableRemindersCommand(ctx, message)
		case domain.BotCommandDisableReminders.String():
			return b.onDisableRemindersCommand(ctx, message)
		default:
			return b.sendUnsupportedResponse(message.ChatID)
		}
	}

	state, err := b.store.GetBotState(ctx, message.UserID)
	if err != nil {
		return err
	}

	switch state.Name {
	case domain.BotStateNameCreateReminder:
		return b.onEnterReminderTextUserMessage(ctx, message)
	case domain.BotStateNameEnterReminAt:
		return b.onEnterRemindAtUserMessage(ctx, message)
	case domain.BotStateNameEditReminder:
		// TODO: implement edit reminder feature
		return errors.ErrUnsupported
	case domain.BotStateNameRemoveReminder:
		return b.onRemoveReminderUserMessage(ctx, message)
	default:
		return b.sendUnsupportedResponse(message.ChatID)
	}
}

func (b *Bot) OnCallbackQuery(ctx context.Context, callback domain.TgCallbackQuery) error {
	if callback.IsButtonClick() {
		switch {
		case strings.HasPrefix(callback.Data, domain.ButtonDataPrefixReminderDone):
			return b.onDoneReminderButton(ctx, callback)
		case strings.HasPrefix(callback.Data, domain.ButtonDataRemoveReminder):
			return b.onRemoveReminderButton(ctx, callback)
		case callback.IsRemindAtButtonClick():
			return b.onRemindAtButton(ctx, callback)
		case strings.HasPrefix(callback.Data, domain.ButtonDataPrefixDelayReminder):
			return b.onDelayReminderButton(ctx, callback)
		case strings.HasPrefix(callback.Data, domain.ButtonDataEditReminder):
			return b.onEditReminderButton(ctx, callback)
		default:
			return b.sendUnsupportedResponse(callback.ChatID)
		}
	}

	return b.sendUnsupportedResponse(callback.ChatID)
}

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
}

func (b *Bot) sendUnsupportedResponse(chatID int64) error {
	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: chatID,
		Text:   fmt.Sprintf("Я не понимаю о чём речь! Пожалуйста, воспользуйся командой %s.", domain.BotCommandHelp),
	})
}

func (b *Bot) createReminder(ctx context.Context, userID, chatID int64, remindAt time.Time) error {
	botState, err := b.store.GetBotState(ctx, userID)
	if err != nil {
		return err
	}

	if botState.Name != domain.BotStateNameEnterReminAt {
		return fmt.Errorf("can't create reminder: invalid bot state: expected [%s], acttual [%s]", domain.BotStateNameEnterReminAt, botState.Name)
	}

	remidner := domain.Reminder{
		ChatID:       chatID,
		UserID:       userID,
		Text:         botState.ReminderText(),
		RemindAt:     remindAt.UTC(),
		Status:       domain.ReminderStatusPending,
		AttemptsLeft: domain.DefaultAttemptsLeft,
	}

	if _, err = b.store.SaveReminder(ctx, remidner); err != nil {
		return err
	}

	botState.Context = nil
	// go to start state
	botState.Name = domain.BotStateNameStart
	if err = b.store.SaveBotState(ctx, botState); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: chatID,
		Text:   fmt.Sprintf(`*%s* я напомню тебе о *%s*!`, domain.MoscowTime(remidner.RemindAt).Format(domain.LayoutRemindAt), remidner.Text),
	})
}

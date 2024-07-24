package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
	"github.com/mezk/tg-reminder/internal/pkg/storage"
)

func (b *Bot) onStartCommand(ctx context.Context, message domain.TgMessage) error {
	user := domain.User{
		ID:     message.UserID,
		Name:   message.UserName,
		Status: domain.UserStatusActive,
	}

	if err := b.store.SaveUser(ctx, user); err != nil {
		switch {
		case errors.Is(err, storage.ErrUserAlreadyExists):
			// client already registered
			if err = b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameStart}); err != nil {
				return err
			}
			return b.responseSender.SendBotResponse(sender.BotResponse{
				ChatID: message.ChatID,
				Text:   fmt.Sprintf("@%s, ранее мы уже начали общение, предлагаю продолжить!", user.Name),
			})
		default:
			return err
		}
	}

	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameStart}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: message.ChatID,
		Text:   fmt.Sprintf("Привет, @%s!\nТеперь ты можешь со мной работать.\nДля справки используй команду /help.", user.Name),
	})
}

func (b *Bot) onHelpCommand(ctx context.Context, message domain.TgMessage) error {
	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameHelp}); err != nil {
		return err
	}

	var helpMsg = fmt.Sprintf(`
		Список доступных команд:
		- %s - cправка
		- %s - начать работу с ботом
		- %s - cоздать напоминание
		- %s - включить напоминания
		- %s - выключить напоминания
		- %s - мои напоминания`,
		domain.BotCommandStart.Markdown(),
		domain.BotCommandHelp.Markdown(),
		domain.BotCommandCreateReminder.Markdown(),
		domain.BotCommandEnableReminders.Markdown(),
		domain.BotCommandDisableReminders.Markdown(),
		domain.BotCommandMyReminders.Markdown(),
	)

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: message.ChatID,
		Text:   helpMsg,
	})
}

func (b *Bot) onCreateReminderCommand(ctx context.Context, message domain.TgMessage) error {
	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameCreateReminder}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{ChatID: message.ChatID, Text: "О чём напомнить?"})
}

func (b *Bot) onMyRemindersCommand(ctx context.Context, message domain.TgMessage) error {
	reminders, err := b.store.GetMyReminders(ctx, message.UserID, message.ChatID)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return b.responseSender.SendBotResponse(sender.BotResponse{
			ChatID: message.ChatID,
			Text:   "У тебя нет напоминаний!",
		})
	}

	var sb strings.Builder
	sb.WriteString("*Вот список твоих активных напоминаний:*")
	sb.WriteString("\n")

	for i, r := range reminders {
		sb.WriteString(r.UserFormat())
		if i+1 < len(reminders) {
			sb.WriteString("\n")
		}
	}

	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameMyReminders}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: message.ChatID,
		Text:   sb.String(),
	}, sender.WithMyRemindersListEditButtons())
}

func (b *Bot) onEnableRemindersCommand(ctx context.Context, message domain.TgMessage) error {
	if err := b.store.SetUserStatus(ctx, message.UserID, domain.UserStatusActive); err != nil {
		return err
	}

	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameEnableReminders}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: message.ChatID,
		Text:   fmt.Sprintf(`Уведомления включены. Для отключения уведомлений воспользуйтесь командой %s.`, domain.BotCommandDisableReminders.Markdown()),
	})
}

func (b *Bot) onDisableRemindersCommand(ctx context.Context, message domain.TgMessage) error {
	if err := b.store.SetUserStatus(ctx, message.UserID, domain.UserStatusInactive); err != nil {
		return err
	}

	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameDisableReminders}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: message.ChatID,
		Text:   fmt.Sprintf("Уведомления отключены. Для включения уведомлений воспользуйтесь командой %s.", domain.BotCommandEnableReminders.Markdown()),
	})
}

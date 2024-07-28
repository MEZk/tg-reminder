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
				Text:   fmt.Sprintf("@%s, ранее мы уже начали общение, предлагаю продолжить %s", user.Name, domain.EmojiWavingHand),
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
		Text:   fmt.Sprintf("*Привет,* @%s %s\n\nТеперь вы можете со мной работать.\nДля справки %s используйте команду %s", user.Name, domain.EmojiWavingHand, domain.EmojiPersonTippingHand, domain.BotCommandHelp.Markdown()),
	})
}

func (b *Bot) onHelpCommand(ctx context.Context, message domain.TgMessage) error {
	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameHelp}); err != nil {
		return err
	}

	var helpMsg = fmt.Sprintf(`
*Список доступных команд*
	• %s — cправка %s
	• %s — начать работу с ботом %s
	• %s — создать напоминание %s
	• %s — включить напоминания %s
	• %s — выключить напоминания %s
	• %s — мои напоминания %s`,
		domain.BotCommandHelp.Markdown(), domain.EmojiPersonTippingHand,
		domain.BotCommandStart.Markdown(), domain.EmojiPlayButton,
		domain.BotCommandCreateReminder.Markdown(), domain.EmojiMemo,
		domain.BotCommandEnableReminders.Markdown(), domain.EmojiBell,
		domain.BotCommandDisableReminders.Markdown(), domain.EmojiBellWithSlash,
		domain.BotCommandMyReminders.Markdown(), domain.EmojiSpiralNotepad,
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

	return b.responseSender.SendBotResponse(sender.BotResponse{ChatID: message.ChatID, Text: fmt.Sprintf("О чём напомнить%s", domain.EmojiQuestionMark)})
}

func (b *Bot) onMyRemindersCommand(ctx context.Context, message domain.TgMessage) error {
	reminders, err := b.store.GetMyReminders(ctx, message.UserID, message.ChatID)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return b.responseSender.SendBotResponse(sender.BotResponse{
			ChatID: message.ChatID,
			Text:   fmt.Sprintf("*У вас нет напоминаний* %s\n\nЧтобы добавить напоминание используйте команду %s", domain.EmojiDisappointedFace, domain.BotCommandCreateReminder.Markdown()),
		})
	}

	const doubleNewLine = "\n\n"

	var sb strings.Builder
	sb.WriteString("*СПИСОК НАПОМИНАНИЙ*")
	sb.WriteString(doubleNewLine)

	for _, r := range reminders {
		sb.WriteString(r.FormatList(timeNowUTC()))
		sb.WriteString(doubleNewLine)
	}

	if err = b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameMyReminders}); err != nil {
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
		Text:   fmt.Sprintf("*Уведомления включены* %s\n\nДля отключения уведомлений используйте команду %s", domain.EmojiBell, domain.BotCommandDisableReminders.Markdown()),
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
		Text:   fmt.Sprintf("*Уведомления отключены* %s\n\nДля включения уведомлений воспользуйтесь командой %s", domain.EmojiBellWithSlash, domain.BotCommandEnableReminders.Markdown()),
	})
}

package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
	"github.com/mezk/tg-reminder/internal/pkg/storage"
)

func (b *Bot) onEnterReminderTextUserMessage(ctx context.Context, message domain.TgMessage) error {
	state := domain.BotState{
		UserID: message.UserID,
		Name:   domain.BotStateNameEnterReminAt,
	}
	state.SetReminderText(message.Text)

	if err := b.store.SaveBotState(ctx, state); err != nil {
		return err
	}

	var text = fmt.Sprintf(`*Когда напомнить?*

Текущая дата и время (Москва): %s.

Введи дату в формате *YYYY-MM-DD HH:mm*.
	
Например, 2024-06-07 11:30 значит, что я пришлю тебе напоминание 7 мая 2024 года в 11:30.

Или выберите опцию ниже:`, domain.MoscowTime(timeNowUTC()).Format(domain.LayoutRemindAt))

	return b.responseSender.SendBotResponse(sender.BotResponse{ChatID: message.ChatID, Text: text}, sender.WithReminderDatesButtons())
}

func (b *Bot) onEnterRemindAtUserMessage(ctx context.Context, message domain.TgMessage) error {
	remindAt, err := message.RemindAt()
	if err != nil {
		return err
	}

	return b.createReminder(ctx, message.UserID, message.ChatID, remindAt)
}

func (b *Bot) onRemoveReminderUserMessage(ctx context.Context, message domain.TgMessage) error {
	reminderID, err := strconv.ParseInt(message.Text, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse reminder id %s: %w", message.Text, err)
	}

	responseMsg := fmt.Sprintf("Напоминание %d удалено!", reminderID)

	if err = b.store.RemoveReminder(ctx, reminderID); err != nil {
		switch {
		case errors.Is(err, storage.ErrReminderNotFound):
			responseMsg = fmt.Sprintf("Напоминание %d не найдено!", reminderID)
		default:
			return err
		}
	}

	// go to start state
	if err = b.store.SaveBotState(ctx, domain.BotState{UserID: message.UserID, Name: domain.BotStateNameStart}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{ChatID: message.ChatID, Text: responseMsg})
}
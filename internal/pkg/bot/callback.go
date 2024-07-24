package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
)

func (b *Bot) onDoneReminderButton(ctx context.Context, callback domain.TgCallbackQuery) error {
	reminderID, err := callback.ReminderID()
	if err != nil {
		return fmt.Errorf("can'p parse reminderID: %w", err)
	}

	if err = b.store.SetReminderStatus(ctx, reminderID, domain.ReminderStatusDone); err != nil {
		return err
	}

	// go to start state
	if err = b.store.SaveBotState(ctx, domain.BotState{UserID: callback.UserID, Name: domain.BotStateNameStart}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{ChatID: callback.ChatID, Text: "Я пометил напоминание как выполненное!"})
}

func (b *Bot) onRemoveReminderButton(ctx context.Context, callback domain.TgCallbackQuery) error {
	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: callback.UserID, Name: domain.BotStateNameRemoveReminder}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: callback.ChatID,
		Text:   "Напиши номер напоминания для удаления.",
	})
}

func (b *Bot) onDelayReminderButton(ctx context.Context, callback domain.TgCallbackQuery) error {
	data, ok := strings.CutPrefix(callback.Data, domain.ButtonDataPrefixDelayReminder)
	if !ok {
		return fmt.Errorf("invalid delay duration format: %s", data)
	}

	reminderID, err := callback.ReminderID()
	if err != nil {
		return fmt.Errorf("can't parse reminderID: %w", err)
	}

	remindAt, err := callback.RemindAt(timeNowUTC())
	if err != nil {
		return fmt.Errorf("can't parse delay: %w", err)
	}

	if err = b.store.DelayReminder(ctx, reminderID, remindAt.In(time.UTC)); err != nil {
		return err
	}

	// go to start state
	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: callback.UserID, Name: domain.BotStateNameStart}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: callback.ChatID,
		Text:   fmt.Sprintf("Я отложил напоминание. Напомню позже *%s*!", remindAt.Format(domain.LayoutRemindAt)),
	})
}

func (b *Bot) onRemindAtButton(ctx context.Context, callback domain.TgCallbackQuery) error {
	remindAt, err := callback.RemindAt(timeNowUTC())
	if err != nil {
		return err
	}

	return b.createReminder(ctx, callback.UserID, callback.ChatID, remindAt)
}

func (b *Bot) onEditReminderButton(ctx context.Context, callback domain.TgCallbackQuery) error {
	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: callback.UserID, Name: domain.BotStateNameEditReminder}); err != nil {
		return err
	}

	return b.responseSender.SendBotResponse(sender.BotResponse{
		ChatID: callback.ChatID,
		Text:   "Напиши номер напоминания для редактирования.",
	})
}

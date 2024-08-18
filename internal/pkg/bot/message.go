package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	log "github.com/go-pkgz/lgr"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
	"github.com/mezk/tg-reminder/internal/pkg/storage"
)

const enterRemindAtFormats = `*Вы можете использовать следующие форматы:*

- в 19:00
- завтра
- завтра в 19:00
- в среду в 15:00
- через час
- через 2 часа
- 30.01.2024 в 11:00
- через месяц
- 2024-08-29 11:30

*Введите дату и время напоминания или выберите опцию ниже:*`

func (b *Bot) onEnterReminderTextUserMessage(ctx context.Context, message domain.TgMessage) error {
	state := domain.BotState{
		UserID: message.UserID,
		Name:   domain.BotStateNameEnterReminAt,
	}
	state.SetReminderText(message.Text)

	if err := b.store.SaveBotState(ctx, state); err != nil {
		return err
	}

	text := fmt.Sprintf("*Когда напомнить %s\n\n*Текущая дата и время (Москва)%s%s\n*%s*\n\n%s",
		domain.EmojiQuestionMark,
		domain.NoBreakSpace, domain.EmojiAlarmClock,
		domain.MoscowTime(timeNowUTC()).Format(domain.LayoutRemindAt),
		enterRemindAtFormats,
	)

	return b.responseSender.SendBotResponse(sender.BotResponse{ChatID: message.ChatID, Text: text}, sender.WithReminderDatesButtons())
}

func (b *Bot) onEnterRemindAtUserMessage(ctx context.Context, message domain.TgMessage) error {
	remindAt, err := message.RemindAt(timeNowUTC())
	if err != nil {
		log.Printf("[WARN] failed to parse remindAt from %s: %s", message.Text, err)

		text := fmt.Sprintf("%s Не удалось понять время из запроса, пожалуйста, попытайтесь его изменить. Время должно быть в будущем.\n\n%s",
			domain.EmojiThinkingFace,
			enterRemindAtFormats,
		)

		return b.responseSender.SendBotResponse(sender.BotResponse{ChatID: message.ChatID, Text: text}, sender.WithReminderDatesButtons())
	}

	return b.createReminder(ctx, message.UserID, message.ChatID, remindAt)
}

func (b *Bot) onRemoveReminderUserMessage(ctx context.Context, message domain.TgMessage) error {
	reminderID, err := strconv.ParseInt(message.Text, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse reminder id %s: %w", message.Text, err)
	}

	responseMsg := fmt.Sprintf("Напоминание %d удалено %s", reminderID, domain.EmojiCrossMark)

	if err = b.store.RemoveReminder(ctx, reminderID); err != nil {
		switch {
		case errors.Is(err, storage.ErrReminderNotFound):
			responseMsg = fmt.Sprintf("Напоминание %d не найдено %s", reminderID, domain.EmojiThinkingFace)
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

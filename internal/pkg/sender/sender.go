package sender

import (
	"fmt"
	"strconv"

	log "github.com/go-pkgz/lgr"
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

// BotAPI - subset of Telegram bot API methods.
type BotAPI interface {
	Send(c tbapi.Chattable) (tbapi.Message, error)
}

type botResponseSender struct {
	botAPI BotAPI
}

// New creates a sender of Telegram's bot responses.
func New(botAPI BotAPI) *botResponseSender {
	return &botResponseSender{botAPI: botAPI}
}

// Send sends a message to telegram as markdown first and if failed - as plain text.
func (s *botResponseSender) SendBotResponse(resp BotResponse, opts ...BotResponseOption) error {
	log.Printf("[DEBUG] bot response - %s", resp)

	for _, opt := range opts {
		opt(&resp)
	}

	tbMsg := tbapi.NewMessage(resp.ChatID, resp.Text)
	tbMsg.ParseMode = tbapi.ModeMarkdown
	tbMsg.DisableWebPagePreview = true
	tbMsg.ReplyToMessageID = int(resp.ReplyToMessageID)
	setReplyMarkup(&tbMsg, resp)

	if err := s.send(tbMsg); err != nil {
		return fmt.Errorf("can't send message to telegram %q: %w", resp.Text, err)
	}

	return nil
}

func (s *botResponseSender) send(tbMsg tbapi.Chattable) error {
	withParseMode := func(tbMsg tbapi.Chattable, parseMode string) tbapi.Chattable {
		switch msg := tbMsg.(type) {
		case tbapi.MessageConfig:
			msg.ParseMode = parseMode
			msg.DisableWebPagePreview = true
			return msg
		default:
			return tbMsg // don't touch other types
		}
	}

	msg := withParseMode(tbMsg, tbapi.ModeMarkdown) // try markdown first
	if _, err := s.botAPI.Send(msg); err != nil {
		log.Printf("[WARN] failed to send message to telegram as markdown, %v", err)

		msg = withParseMode(tbMsg, "") // try plain text
		if _, err = s.botAPI.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

func setReplyMarkup(tbMsg *tbapi.MessageConfig, resp BotResponse) {
	if resp.showMyReminderListEditButtons {
		tbMsg.ReplyMarkup = tbapi.NewInlineKeyboardMarkup(
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData("Редактировать", domain.ButtonDataEditReminder),
				tbapi.NewInlineKeyboardButtonData("Удалить", domain.ButtonDataRemoveReminder),
			),
		)
	}

	if resp.showReminderDatesButtons {
		tbMsg.ReplyMarkup = tbapi.NewInlineKeyboardMarkup(
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData("11:30", domain.ButtonDataPrefixRemindAtTime+"11:30"),
				tbapi.NewInlineKeyboardButtonData("14:30", domain.ButtonDataPrefixRemindAtTime+"14:30"),
				tbapi.NewInlineKeyboardButtonData("19:30", domain.ButtonDataPrefixRemindAtTime+"19:30"),
				tbapi.NewInlineKeyboardButtonData("20:30", domain.ButtonDataPrefixRemindAtTime+"20:30"),
			),
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData("30 мин", domain.ButtonDataPrefixRemindAtDuration+"30m"),
				tbapi.NewInlineKeyboardButtonData("80 мин", domain.ButtonDataPrefixRemindAtDuration+"80m"),
				tbapi.NewInlineKeyboardButtonData("1 день", domain.ButtonDataPrefixRemindAtDuration+"24h"),
				tbapi.NewInlineKeyboardButtonData("1 месяц", domain.ButtonDataPrefixRemindAtDuration+"730h"),
			),
		)
	}

	if resp.showReminderDoneButtons {
		reminderID := strconv.FormatInt(resp.reminderID, 10)

		tbMsg.ReplyMarkup = tbapi.NewInlineKeyboardMarkup(
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData("Готово", domain.ButtonDataPrefixReminderDone+reminderID),
			),
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData("Отложить на 15 мин", domain.ButtonDataPrefixDelayReminder+reminderID+"/15m"),
				tbapi.NewInlineKeyboardButtonData("Отложить на 30 мин", domain.ButtonDataPrefixDelayReminder+reminderID+"/30m"),
			),
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData("Отложить на 1 час", domain.ButtonDataPrefixDelayReminder+reminderID+"/1h"),
				tbapi.NewInlineKeyboardButtonData("Отложить на 1 день", domain.ButtonDataPrefixDelayReminder+reminderID+"/24h"),
			),
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData("Отложить на 1 неделю", domain.ButtonDataPrefixDelayReminder+reminderID+"/168h"),
			),
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData("Отложить на 1 месяц", domain.ButtonDataPrefixDelayReminder+reminderID+"/730h"),
			),
		)
	}
}
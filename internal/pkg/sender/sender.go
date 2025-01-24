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

// BotResponseSender - sender which is able to send bot response to user.
type BotResponseSender struct {
	botAPI BotAPI
}

// New - creates a sender of Telegram's bot responses.
// TODO: реализовать установку команд https://github.com/go-telegram-bot-api/telegram-bot-api/blob/4126fa611266940425a9dfd37e0c92ba47881718/bot_test.go#L959
func New(botAPI BotAPI) *BotResponseSender {
	return &BotResponseSender{botAPI: botAPI}
}

// SendBotResponse - sends a message to telegram as markdown first and if failed - as plain text.
func (s *BotResponseSender) SendBotResponse(resp BotResponse, opts ...BotResponseOption) error {
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

func (s *BotResponseSender) send(tbMsg tbapi.Chattable) error {
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

const (
	buttonTextEditReminder     = domain.EmojiMemo + " Редактировать"
	buttonTextRemoveReminder   = domain.EmojiCrossMark + " Удалить"
	buttonTextReminderDone     = domain.EmojiWhiteHeavyCheckMark + " Готово"
	buttonText15Min            = domain.EmojiCounterclockwiseArrowsButton + " 30 мин."
	buttonText30Min            = domain.EmojiCounterclockwiseArrowsButton + " 80 мин."
	buttonText1Hour            = domain.EmojiCounterclockwiseArrowsButton + " 3 час."
	buttonText1Day             = domain.EmojiCounterclockwiseArrowsButton + " 1 ден."
	buttonText1Week            = domain.EmojiCounterclockwiseArrowsButton + " 1 нед."
	buttonText1Month           = domain.EmojiCounterclockwiseArrowsButton + " 1 мес."
	buttomTextEditRemidnerText = "Текст напоминания"
	buttomTextEditRemidnerDate = "Дату напоминания"
)

func setReplyMarkup(tbMsg *tbapi.MessageConfig, resp BotResponse) {
	if resp.showMyReminderListEditButtons {
		tbMsg.ReplyMarkup = tbapi.NewInlineKeyboardMarkup(
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData(buttonTextEditReminder, domain.ButtonDataEditReminder),
				tbapi.NewInlineKeyboardButtonData(buttonTextRemoveReminder, domain.ButtonDataRemoveReminder),
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
				tbapi.NewInlineKeyboardButtonData(buttonText15Min, domain.ButtonDataPrefixDelayReminder+reminderID+"/30m"),
				tbapi.NewInlineKeyboardButtonData(buttonText30Min, domain.ButtonDataPrefixDelayReminder+reminderID+"/80m"),
				tbapi.NewInlineKeyboardButtonData(buttonText1Hour, domain.ButtonDataPrefixDelayReminder+reminderID+"/3h"),
			),
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData(buttonText1Day, domain.ButtonDataPrefixDelayReminder+reminderID+"/24h"),
				tbapi.NewInlineKeyboardButtonData(buttonText1Week, domain.ButtonDataPrefixDelayReminder+reminderID+"/168h"),
				tbapi.NewInlineKeyboardButtonData(buttonText1Month, domain.ButtonDataPrefixDelayReminder+reminderID+"/730h"),
			),
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData(buttonTextReminderDone, domain.ButtonDataPrefixReminderDone+reminderID),
			),
		)
	}

	if resp.showReminderEditButtons {
		tbMsg.ReplyMarkup = tbapi.NewInlineKeyboardMarkup(
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData(buttomTextEditRemidnerText, domain.ButtonDataEditReminderText),
			),
			tbapi.NewInlineKeyboardRow(
				tbapi.NewInlineKeyboardButtonData(buttomTextEditRemidnerDate, domain.ButtonDataEditReminderDate),
			),
		)
	}
}

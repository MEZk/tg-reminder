package v2

import (
	"context"
	"fmt"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/markusmobius/go-dateparser"
	"github.com/markusmobius/go-dateparser/date"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

const layoutTimeOnly = "15:04"

func (tb *tgBot) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := getUserID(update)
	text := update.Message.Text

	switch state := tb.state.Current(userID); state {
	case stateCreateReminderAskText:
		reminderTextLen := len([]rune(text))
		const maxReminderLen = 500
		if reminderTextLen > maxReminderLen {
			tb.sendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   fmt.Sprintf("Длина напоминания равна <b>%d</b>, что превышает <b>%d</b> символов. Пожалуйста, сократите текст напоминания.", reminderTextLen, maxReminderLen),
			})
		}

		tb.state.Set(userID, stateKeyReminderText, text)
		tb.state.Transition(userID, stateCreateReminderAskDate, ctx, userID)

		tb.askReminderDate(ctx, userID)
	case stateCreateReminderAskDate:
		reminderDate, err := getReminderDateFromUserMsg(text, timeNowUTC())
		if err != nil {
			log.Printf("[ERROR] can't get reminder date from user message: %v", err)

			tb.sendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   "🤔 Не удалось понять время из запроса, пожалуйста, попытайтесь его изменить. Время должно быть в будущем.",
			})
			return
		}

		tb.createReminder(ctx, reminderDate, userID)
	case stateCreateReminderAskTime:
		d, ok := tb.state.Get(userID, stateKeyReminderDateOnly)
		if !ok {
			log.Printf("[ERROR] invalid user %d state: reminder date not found", userID)
			return
		}

		// TODO: перепроверить часовой пояс в БД и при выоводе, что-то тут странное
		date := domain.MoscowTime(d.(time.Time))

		t, err := time.Parse(layoutTimeOnly, text)
		if err != nil {
			log.Printf("[ERROR] can't get reminder time from user message: %v", err)

			tb.sendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   "🤔 Не удалось понять время напоминания из запроса, пожалуйста, попытайтесь его изменить.",
			})

			tb.onCalendarReminderDateSelected(ctx, tb.api, models.MaybeInaccessibleMessage{Message: update.Message}, date)
			return
		}

		reminderDate := time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())

		tb.createReminder(ctx, reminderDate.UTC(), userID)
	default:
		log.Printf("[ERROR] fail to handle user %d message: invalid state '%s'", userID, state)
	}
}

func getReminderDateFromUserMsg(msg string, now time.Time) (time.Time, error) {
	now = domain.MoscowTime(now)

	remindAt, err := time.ParseInLocation(domain.LayoutRemindAt, msg, now.Location())
	if err != nil {
		log.Printf("[DEBUG] can't parse (time.ParseInLocation) remindAt %s: %s\n", msg, err)

		var remindAtDate date.Date
		if remindAtDate, err = dateparser.Parse(&dateparser.Configuration{ // TODO: здесь в БД записывается неверная дата
			CurrentTime:         now,
			Locales:             []string{"ru"},
			PreferredDateSource: dateparser.Future,
		}, msg); err != nil {
			return time.Time{}, err
		}

		remindAt = remindAtDate.Time
	}

	return remindAt.Truncate(1 * time.Minute), nil
}

package v2

import (
	"context"
	"fmt"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/datepicker"
	"github.com/go-telegram/ui/keyboard/inline"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

const commandCreateReminder = "/create_reminder"

const (
	btnCalendar = "btn_calendar"
	btnCancel   = "btn_cancel"
	btn11_30    = "btn_11_30"
	btn14_30    = "btn_14_30"
	btn19_30    = "btn_19_30"
	btn20_30    = "btn_20_30"
	btn30min    = "btn_30min"
	btn80min    = "btn_80min"
	btn1day     = "btn_1day"
	btn1month   = "btn_1month"
)

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
}

func (tb *tgBot) createReminderHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := getUserID(update)

	tb.sendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   "<b>Отправьте мне текст напоминания</b>",
	})

	tb.state.Transition(userID, stateCreateReminderAskText)
}

func (tb *tgBot) askReminderDate(ctx context.Context, userID int64) {
	kb := inline.New(tb.api).
		Row().
		Button("11:30", []byte(btn11_30), tb.onKeyboardReminderTimeSelected).
		Button("14:30", []byte(btn14_30), tb.onKeyboardReminderTimeSelected).
		Button("19:30", []byte(btn19_30), tb.onKeyboardReminderTimeSelected).
		Button("20:30", []byte(btn20_30), tb.onKeyboardReminderTimeSelected).
		Row().
		Button("30 мин", []byte(btn30min), tb.onKeyboardReminderTimeSelected).
		Button("80 мин", []byte(btn80min), tb.onKeyboardReminderTimeSelected).
		Button("1 день", []byte(btn1day), tb.onKeyboardReminderTimeSelected).
		Button("1 месяц", []byte(btn1month), tb.onKeyboardReminderTimeSelected).
		Row().
		Button("Календарь", []byte(btnCalendar), tb.onKeyboardCalendarSelected).
		Button("Отмена", []byte(btnCancel), tb.onKeyboardCancelSelected)

	tb.sendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text: fmt.Sprintf("Когда напомнить ❓\n\n<b>Сегодня %s</b> (Москва).\n\n<b>Вы можете использовать следующие форматы для ввода даты и времени напоминания:</b>\n\n• в 19:00\n• завтра\n• завтра в 19:00\n• в среду в 15:00\n• через час\n• через 2 часа\n• 30.01.2024 в 11:00\n• через месяц\n• 2024-08-29 11:30\n\n<b>Введите дату и время напоминания или выберите опцию ниже:</b>\n",
			domain.MoscowTime(timeNowUTC()).Format(domain.LayoutRemindAt)),
		ReplyMarkup: kb,
	})
}

func (tb *tgBot) onKeyboardReminderTimeSelected(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if mes.Message == nil {
		return
	}

	userID := mes.Message.Chat.ID

	reminderDate, err := getReminderDateFromButtonData(data, timeNowUTC())
	if err != nil {
		log.Printf("[ERROR] can't get reminder date from button data: %v", err)
		tb.state.Transition(userID, stateDefault)
		return
	}

	tb.createReminder(ctx, reminderDate, userID)
}

func (tb *tgBot) createReminder(ctx context.Context, reminderDate time.Time, userID int64) {
	val, ok := tb.state.Get(userID, stateKeyReminderText)
	if !ok {
		log.Printf("[ERROR] can't get reminder text by key '%s': not found", stateKeyReminderText)
		tb.state.Transition(userID, stateDefault)
		return
	}

	reminderText := val.(string)

	if _, err := tb.store.SaveReminder(ctx, domain.Reminder{
		ChatID:       userID,
		UserID:       userID,
		Text:         reminderText,
		RemindAt:     reminderDate.UTC(),
		Status:       domain.ReminderStatusPending,
		AttemptsLeft: domain.DefaultAttemptsLeft,
	}); err != nil {
		log.Printf("[ERROR] can't save reminder: %v", err)

		tb.state.Transition(userID, stateDefault)
		return
	}

	tb.sendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   fmt.Sprintf("<b>Создано напоминание</b>\n\n<i>%s</i>\n\nНапомнить: %s", reminderText, domain.MoscowTime(reminderDate).Format(domain.LayoutRemindAt)),
	})

	tb.state.Transition(userID, stateDefault)
}

func (tb *tgBot) onKeyboardCancelSelected(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if mes.Message == nil || mes.Message.From == nil {
		return
	}

	tb.state.Transition(mes.Message.From.ID, stateDefault)
}

func getReminderDateFromButtonData(data []byte, now time.Time) (time.Time, error) {
	nowMSK := domain.MoscowTime(now)

	var reminderDate time.Time
	switch string(data) {
	case btn11_30:
		reminderDate = time.Date(nowMSK.Year(), nowMSK.Month(), nowMSK.Day(), 11, 30, 0, 0, nowMSK.Location())
	case btn14_30:
		reminderDate = time.Date(nowMSK.Year(), nowMSK.Month(), nowMSK.Day(), 14, 30, 0, 0, nowMSK.Location())
	case btn19_30:
		reminderDate = time.Date(nowMSK.Year(), nowMSK.Month(), nowMSK.Day(), 19, 30, 0, 0, nowMSK.Location())
	case btn20_30:
		reminderDate = time.Date(nowMSK.Year(), nowMSK.Month(), nowMSK.Day(), 20, 30, 0, 0, nowMSK.Location())
	case btn30min:
		reminderDate = nowMSK.Add(30 * time.Minute)
	case btn80min:
		reminderDate = nowMSK.Add(30 * time.Minute)
	case btn1day:
		reminderDate = nowMSK.Add(24 * time.Hour)
	case btn1month:
		reminderDate = nowMSK.AddDate(0, 1, 0)
	default:
		return time.Time{}, fmt.Errorf("unknown button data: %s", string(data))
	}

	if reminderDate.Before(nowMSK) {
		reminderDate = reminderDate.AddDate(0, 0, 1)
	}

	return reminderDate.Truncate(1 * time.Minute).UTC(), nil
}

// TODO: все, что ниже - доделать, перепроверить

func (tb *tgBot) onKeyboardCalendarSelected(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	if mes.Message == nil {
		return
	}

	userID := mes.Message.Chat.ID
	// if tb.state.Current(userID) != stateWaitReminderDate {
	// 	return
	// }

	log.Printf("[DEBUG] onReminderDateKeyboardCalendar: user %d", userID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      userID,
		Text:        "Выберите дату напоминания",
		ReplyMarkup: datepicker.New(b, tb.onCalendarReminderDateSelected, datepicker.Language("ru")),
	})
}

func (tb *tgBot) onCalendarReminderDateSelected(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, date time.Time) {
	if mes.Message == nil {
		return
	}

	userID := mes.Message.Chat.ID

	tb.state.Set(userID, stateKeyReminderDateOnly, date)

	kb := inline.New(tb.api).
		Row().
		Button("11:30", []byte(btn11_30), tb.onReminderDateKeyboardTimeSelected).
		Button("14:30", []byte(btn14_30), tb.onReminderDateKeyboardTimeSelected).
		Row().
		Button("19:30", []byte(btn19_30), tb.onReminderDateKeyboardTimeSelected).
		Button("20:30", []byte(btn20_30), tb.onReminderDateKeyboardTimeSelected).
		Row().
		Button("Отмена", []byte(btnCancel), tb.onReminderDateKeyboardCancel)

	tb.sendMessage(ctx, &bot.SendMessageParams{
		ChatID:      userID,
		Text:        "<b>Выберите время напоминания или введите его в формате HH:mm, например, 11:35</b> ",
		ReplyMarkup: kb,
	})

	tb.state.Transition(userID, stateCreateReminderAskTime)
}

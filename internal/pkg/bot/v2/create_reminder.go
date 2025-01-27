package v2

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/markusmobius/go-dateparser"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	tele "gopkg.in/telebot.v4"
)

func (b *tgBot) onCreateReminderCmd(c tele.Context) error {
	msg, err := c.Bot().Send(c.Recipient(), "<b>О чём напомнить❓</b>")
	if err != nil {
		return err
	}

	state := domain.BotState{
		UserID: c.Sender().ID,
		Name:   domain.BotStateNameCreateReminder,
		Context: &domain.BotStateContext{
			PrevMessageIDs: []int{c.Message().ID, msg.ID},
		},
	}

	return b.store.SaveBotState(context.TODO(), state)
}

var (
	btn11_30 = tele.Btn{Unique: "btn11_30", Text: "11:30", Data: "11:30"}
	btn14_30 = tele.Btn{Unique: "btn14_30", Text: "14:30", Data: "14:30"}
	btn19_30 = tele.Btn{Unique: "btn19_30", Text: "19:30", Data: "19:30"}
	btn20_30 = tele.Btn{Unique: "btn20_30", Text: "20:30", Data: "20:30"}

	btn30m  = tele.Btn{Unique: "btn30m", Text: "30 мин", Data: "30m"}
	btn80m  = tele.Btn{Unique: "btn80m", Text: "80 мин", Data: "80m"}
	btn1day = tele.Btn{Unique: "btn1day", Text: "1 день", Data: "24h"}
	btn1mon = tele.Btn{Unique: "btn1mon", Text: "1 месяц", Data: "730h"}
)

var timeNowMSK = func() time.Time {
	return domain.MoscowTime(time.Now().UTC())
}

func (b *tgBot) onReminderTextReceived(ctx context.Context, c tele.Context, state domain.BotState) error {
	text := fmt.Sprintf(`
<b>Когда напомнить❓</b>

Сегодня %s (Москва) ⏰

Вы можете использовать следующие форматы:

• в 19:00
• завтра
• завтра в 19:00
• в среду в 15:00
• через час
• через 2 часа
• 30.01.2024 в 11:00
• через месяц
• 2024-08-29 11:30

Введите дату и время напоминания или выберите опцию ниже
`, timeNowMSK().Format("02.01.2006 15:04"))

	selector := tele.ReplyMarkup{}
	selector.Inline(
		selector.Row(btn11_30, btn14_30, btn19_30, btn20_30),
		selector.Row(btn30m, btn80m, btn1day, btn1mon),
	)

	msg, err := c.Bot().Send(c.Recipient(), text, &selector)
	if err != nil {
		return err
	}

	state.AppendPrevMessageID(c.Message().ID, msg.ID)
	state.SetReminderText(strings.TrimSpace(c.Text()))
	state.Name = domain.BotStateNameEnterReminAt

	return b.store.SaveBotState(ctx, state)
}

func (b *tgBot) onReminderDateBtn(c tele.Context) error {
	ctx := context.TODO()
	userID := c.Sender().ID

	state, err := b.store.GetBotState(ctx, userID)
	if err != nil {
		return err
	}

	if state.Name != domain.BotStateNameEnterReminAt {
		return nil
	}

	remindAt, err := parseRemidnerDate(c.Data(), timeNowMSK(), false)
	if err != nil {
		return fmt.Errorf("failed to get reminder date for user %d from button data: %w", userID, err)
	}

	return b.createReminder(ctx, c, state, userID, remindAt)
}

func (b *tgBot) onRemindDateReceived(ctx context.Context, c tele.Context, state domain.BotState) error {
	userID := c.Sender().ID

	reminderDate, err := parseRemidnerDate(c.Text(), timeNowMSK(), true)
	if err != nil {
		log.Printf("[ERROR] can't get reminder date from user message: %v", err)
		return c.Send("🤔 Не удалось понять время из запроса, пожалуйста, попытайтесь его изменить. Время должно быть в будущем.")
	}

	return b.createReminder(ctx, c, state, userID, reminderDate.UTC())
}

func (b *tgBot) createReminder(ctx context.Context, c tele.Context, state domain.BotState, userID int64, remindAt time.Time) error {
	remidner := domain.Reminder{
		ChatID:       userID,
		UserID:       userID,
		Text:         state.ReminderText(),
		RemindAt:     remindAt.UTC(),
		Status:       domain.ReminderStatusPending,
		AttemptsLeft: domain.DefaultAttemptsLeft,
	}

	if _, err := b.store.SaveReminder(ctx, remidner); err != nil {
		return err
	}

	prevMsgIDs := state.GetPrevMessageIDs()
	if len(prevMsgIDs) > 0 {
		delMessages := make([]tele.Editable, 0, len(prevMsgIDs))
		for _, id := range prevMsgIDs {
			delMessages = append(delMessages, &tele.Message{ID: id, Chat: &tele.Chat{ID: userID}})
		}

		if err := c.Bot().DeleteMany(delMessages); err != nil {
			log.Printf("[ERROR] failed to delete previous messages: %v", err)
		}
	}

	if err := c.Send(fmt.Sprintf(`<b>%s</b> я напомню тебе о <b>%s</b> `, domain.MoscowTime(remidner.RemindAt).Format("02.01.2006 в 15:04"), remidner.Text)); err != nil {
		log.Printf("[ERROR] failed to send message: %v", err)
	}

	state.Context = nil
	state.Name = domain.BotStateNameStart
	return b.store.SaveBotState(ctx, state)
}

func parseRemidnerDate(text string, now time.Time, isUserInput bool) (time.Time, error) {
	var reminderDate time.Time

	if isUserInput || strings.Contains(text, ":") {
		date, err := dateparser.Parse(&dateparser.Configuration{
			CurrentTime:         now,
			Locales:             []string{"ru"},
			PreferredDateSource: dateparser.Future,
		}, text, "2006-01-02 15:04")
		if err != nil {
			return time.Time{}, err
		}

		reminderDate = date.Time
	} else {
		d, err := time.ParseDuration(text)
		if err != nil {
			return time.Time{}, err
		}

		reminderDate = now.Add(d)
	}

	return reminderDate, nil
}

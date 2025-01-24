package v2

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/paginator"
)

const commandMyReminders = "/my_reminders"

func (tb *tgBot) myRemindersHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := getUserID(update)

	reminders, err := tb.store.GetMyReminders(ctx, userID, userID)
	if err != nil {
		log.Printf("[ERROR] can't get user %d reminders: %v", userID, err)
		return
	}

	if len(reminders) == 0 {
		tb.sendMessage(ctx, &bot.SendMessageParams{
			ChatID: userID,
			Text:   fmt.Sprintf("<b>У вас нет напоминаний</b>\n\nЧтобы добавить напоминание используйте команду /create_reminder."),
		})
	}

	const doubleNewLine = "\n\n"

	data := make([]string, 0, len(reminders))
	for _, r := range reminders {
		data = append(data, formatReminder(r, timeNowUTC()))
	}

	p := paginator.New(tb.api, data,
		paginator.PerPage(3),
		paginator.Separator(doubleNewLine),
		paginator.WithCloseButton("Закрыть"),
		paginator.WithEditButton("Редактировать", tb.onEditReminderButton),
		paginator.WithRemoveButton("Удалить", tb.onRemoveReminderButton),
		paginator.WithoutEmptyButtons(),
	)

	if _, err = p.Show(ctx, b, userID); err != nil {
		log.Printf("[ERROR] can't show user %d reminders: %v", userID, err)
	}
}

func (tb *tgBot) onEditReminderButton(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := getUserID(update)

	tb.sendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   "Введите 🆔 напоминания для редактирования.",
	})

	// tb.state.Transition(userID, stateEditReminderAskID)
}

func (tb *tgBot) onRemoveReminderButton(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := getUserID(update)

	tb.sendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   "Введите 🆔 напоминания для удаления.",
	})

	// tb.state.Transition(userID, stateRemoveReminderAskID)
}

func formatReminder(r domain.Reminder, now time.Time) string {
	var (
		remindAtMSK         = domain.MoscowTime(r.RemindAt)
		nYear, nMonth, nDay = domain.MoscowTime(now).Date()
		rYear, rMonth, rDay = remindAtMSK.Date()
		timeOnly            = remindAtMSK.Format(layoutTimeOnly)
	)

	var sb strings.Builder
	sb.WriteString("✅ <b>")
	sb.WriteString(r.Text)
	sb.WriteString("❗</b>")

	if nYear == rYear && nMonth == rMonth && nDay == rDay { // today
		sb.WriteString("\n⏰ Сегодня ")
		sb.WriteString(timeOnly)
	} else {
		sb.WriteString("\n⏰ ")
		sb.WriteString(strconv.Itoa(remindAtMSK.Day()))
		sb.WriteRune(' ')
		sb.WriteString(getRussianMonth(remindAtMSK.Month()))
		sb.WriteRune(' ')
		sb.WriteString(timeOnly)
	}

	sb.WriteString("\n🆔 ")
	sb.WriteString(strconv.FormatInt(r.ID, 10))

	return sb.String()
}

func getRussianMonth(m time.Month) string {
	switch m {
	case time.January:
		return "янв."
	case time.February:
		return "фев."
	case time.March:
		return "мар."
	case time.April:
		return "апр."
	case time.May:
		return "мая"
	case time.June:
		return "июн."
	case time.July:
		return "июл."
	case time.August:
		return "авг."
	case time.September:
		return "сент."
	case time.October:
		return "окт."
	case time.November:
		return "нояб."
	case time.December:
		return "дек."
	default:
		return ""
	}
}

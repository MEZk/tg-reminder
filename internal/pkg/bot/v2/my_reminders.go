package v2

import (
	"context"
	"fmt"
	"strings"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	tele "gopkg.in/telebot.v4"
)

var (
	btnEditReminder   = tele.Btn{Unique: "btn_remove_reminder", Text: "Удалить ❌", Data: "btn_remove_reminder"}
	btnRemoveReminder = tele.Btn{Unique: "btn_edit_reminder", Text: "Редактировать 📝", Data: "btn_edit_reminder"}
)

// TODO: implement pagination
func (b *tgBot) onMyRemindersCmd(c tele.Context) error {
	ctx := context.TODO()
	userID := c.Sender().ID

	reminders, err := b.store.GetMyReminders(ctx, userID, c.Chat().ID)
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		return c.Send(fmt.Sprintf("<b>У вас нет напоминаний</b> 😞\n\nЧтобы добавить напоминание используйте команду %s", cmdCreateReminder))
	}

	var sb strings.Builder
	sb.WriteString("<b>СПИСОК НАПОМИНАНИЙ</b>\n\n")

	for _, r := range reminders {
		sb.WriteString(r.Text) // TODO: format reminder
		sb.WriteString("\n\n")
	}

	selector := tele.ReplyMarkup{ResizeKeyboard: true}
	selector.Inline(selector.Row(btnRemoveReminder, btnEditReminder))

	if _, err = c.Bot().Send(c.Recipient(), sb.String(), &selector); err != nil {
		return err
	}

	return b.store.SaveBotState(ctx, domain.BotState{UserID: userID, Name: domain.BotStateNameMyReminders})
}

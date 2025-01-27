package v2

import (
	"context"
	"fmt"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	tele "gopkg.in/telebot.v4"
)

var helpText = fmt.Sprintf(`
<b>Список доступных команд</b>
	• %s — cправка 💁🏻
	• %s — начать работу с ботом ▶️
	• %s — создать напоминание 📝
	• %s — мои напоминания 🔔`,
	cmdHelp,
	cmdStart,
	cmdCreateReminder,
	cmdMyReminders,
)

func (b *tgBot) onHelpCmd(c tele.Context) error {
	if err := b.store.SaveBotState(context.TODO(), domain.BotState{UserID: c.Sender().ID, Name: domain.BotStateNameHelp}); err != nil {
		return err
	}

	return c.Send(helpText)
}

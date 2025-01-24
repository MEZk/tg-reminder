package v2

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const commandHelp = "/help"

var helpMessage = fmt.Sprintf("<b>Список доступных команд</b>\n• %s — cправка 💁\n• %s — начать работу с ботом ▶️\n• %s — создать напоминание 📝\n• %s — мои напоминания 🗒️",
	commandHelp, commandStart, commandCreateReminder, commandMyReminders)

func (tb *tgBot) helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := getUserID(update)

	tb.sendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   helpMessage,
	})

	tb.state.Transition(userID, stateDefault, userID)
}

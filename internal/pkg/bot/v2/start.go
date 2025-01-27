package v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/storage"
	tele "gopkg.in/telebot.v4"
)

func (b *tgBot) onStartCmd(c tele.Context) error {
	sender := c.Sender()
	user := domain.User{
		ID:     sender.ID,
		Name:   sender.Username,
		Status: domain.UserStatusActive,
	}

	ctx := context.TODO()

	if err := b.store.SaveUser(ctx, user); err != nil {
		switch {
		case errors.Is(err, storage.ErrUserAlreadyExists):
			if err = b.store.SaveBotState(ctx, domain.BotState{UserID: user.ID, Name: domain.BotStateNameStart}); err != nil {
				return err
			}
			return c.Send(fmt.Sprintf("@%s, ты уже отправлял мне команду %s ранее!", user.Name, cmdStart))
		default:
			return err
		}
	}

	if err := b.store.SaveBotState(ctx, domain.BotState{UserID: user.ID, Name: domain.BotStateNameStart}); err != nil {
		return err
	}

	return c.Send(fmt.Sprintf("<b>Привет,</b> @%s 👋🏻\n%s", user.Name, helpText))
}

package v2

import (
	"context"
	"errors"
	"fmt"

	log "github.com/go-pkgz/lgr"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/storage"
)

const commandStart = "/start"

func (tb *tgBot) startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := getUserID(update)

	if state := tb.state.Current(userID); state != stateDefault {
		log.Printf("[WARN] can't register user %d: invalid state '%s'", userID, state)
		return
	}

	username := getUsername(update)

	if err := tb.store.SaveUser(ctx, domain.User{
		ID:     userID,
		Name:   username,
		Status: domain.UserStatusActive,
	}); err != nil {
		switch {
		case errors.Is(err, storage.ErrUserAlreadyExists):
			tb.sendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   fmt.Sprintf("Ранее вы уже отправляли команду %s.\nДля справки используйте команду %s.", commandStart, commandHelp),
			})
		default:
			log.Printf("[ERROR] can't register user %d: %v", userID, err)
		}

		return
	}

	log.Printf("[INFO] registered user %d", userID)

	tb.sendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   fmt.Sprintf("<b>👋 Добро пожаловать, @%s!</b>\n\n%s", username, helpMessage),
	})
}

func getUserID(u *models.Update) int64 {
	if u == nil {
		return -1
	}

	if u.Message != nil && u.Message.From != nil {
		return u.Message.From.ID
	}

	if u.CallbackQuery != nil {
		return u.CallbackQuery.From.ID
	}

	return -1
}

func getUsername(u *models.Update) string {
	if u == nil {
		return ""
	}

	if u.Message != nil && u.Message.From != nil {
		return u.Message.From.Username
	}

	if u.CallbackQuery != nil {
		return u.CallbackQuery.From.Username
	}

	return ""
}

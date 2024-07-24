package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/go-pkgz/lgr"
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

// BotAPI - subset of Telegram bot API methods.
type BotAPI interface {
	GetUpdatesChan(config tbapi.UpdateConfig) tbapi.UpdatesChannel
}

// UpdateReceiver - receiver of Telegram updates.
type UpdateReceiver interface {
	OnMessage(ctx context.Context, message domain.TgMessage) error
	OnCallbackQuery(ctx context.Context, callback domain.TgCallbackQuery) error
}

// Listener - listener which listens to updates from Telegram.
type Listener struct {
	botAPI         BotAPI
	updateReceiver UpdateReceiver
}

// New creates a new [Listener].
func New(botAPI BotAPI, updateReceiver UpdateReceiver) *Listener {
	return &Listener{botAPI: botAPI, updateReceiver: updateReceiver}
}

// Listen - listens to Telegram updates in a loop.
// Blocks until ctx.Err.
func (l *Listener) Listen(ctx context.Context) error {
	log.Printf("[INFO] start telegram updates Listener")

	cfg := tbapi.NewUpdate(0)
	cfg.Timeout = 60

	updates := l.botAPI.GetUpdatesChan(cfg)

	for {
		select {

		case <-ctx.Done():
			return ctx.Err()

		case update, ok := <-updates:
			if !ok {
				return fmt.Errorf("telegram updates chan closed")
			}

			if err := l.processUpdate(ctx, update); err != nil {
				log.Printf("[WARN] failed to process update: %v", err)
				continue
			}
		}
	}
}

func (l *Listener) processUpdate(ctx context.Context, update tbapi.Update) error {
	if update.Message == nil && update.CallbackQuery == nil {
		return nil
	}

	msgJSON, err := json.Marshal(update.Message)
	if err != nil {
		return fmt.Errorf("failed to marshal update.Message to json: %w", err)
	}

	log.Printf("[DEBUG] process update.Message: %s", string(msgJSON))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch {
	case update.Message != nil && update.Message.Text != "":
		message := transformMessage(update.Message)
		if err = l.updateReceiver.OnMessage(ctx, message); err != nil {
			return fmt.Errorf("failed to handle msg (%s): %w", message, err)
		}
	case update.CallbackData() != "":
		callbackQuery := transformCallbackQuery(update.CallbackQuery)
		if err = l.updateReceiver.OnCallbackQuery(ctx, callbackQuery); err != nil {
			return fmt.Errorf("failed to handle callback query (%s): %w", callbackQuery, err)
		}
	default:
		// pass: not interesting in other updates
	}

	return nil
}

func transformMessage(message *tbapi.Message) domain.TgMessage {
	var res domain.TgMessage

	if message != nil {
		if message.Chat != nil {
			res.ChatID = message.Chat.ID
		}

		if message.From != nil {
			res.UserID = message.From.ID
			res.UserName = message.From.UserName
		}

		res.Text = strings.TrimSpace(message.Text)
	}

	return res
}

func transformCallbackQuery(callback *tbapi.CallbackQuery) domain.TgCallbackQuery {
	var res domain.TgCallbackQuery

	if callback != nil {
		res.UserID = callback.From.ID
		res.UserName = callback.From.UserName

		if callback.Message != nil && callback.Message.Chat != nil {
			res.ChatID = callback.Message.Chat.ID
		}

		res.Data = strings.TrimSpace(callback.Data)
	}

	return res
}

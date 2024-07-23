package listener

import (
	"context"
	"errors"
	"testing"
	"time"

	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/listener/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_listener_Listen(t *testing.T) {
	t.Parallel()

	t.Run("success: message", func(t *testing.T) {
		t.Parallel()

		botAPIMock := mocks.BotAPIMock{
			GetUpdatesChanFunc: func(config tbapi.UpdateConfig) tbapi.UpdatesChannel {
				cfg := tbapi.NewUpdate(0)
				cfg.Timeout = 60
				assert.Equal(t, cfg, config)

				ch := make(chan tbapi.Update, 1)
				ch <- tbapi.Update{
					Message: &tbapi.Message{
						MessageID: 13246,
						Text:      "winds",
						From: &tbapi.User{
							ID:       2,
							UserName: "Nirav Martini",
						},
						Chat: &tbapi.Chat{
							ID: 1,
						},
					},
				}
				return ch
			},
		}
		updateReceiverMock := mocks.UpdateReceiverMock{
			OnMessageFunc: func(_ context.Context, message domain.TgMessage) error {
				assert.Equal(t, domain.TgMessage{
					ChatID:   1,
					UserID:   2,
					UserName: "Nirav Martini",
					Text:     "winds",
				}, message)
				return nil
			},
		}

		listenerImpl := New(&botAPIMock, &updateReceiverMock)

		ctx, cancel := context.WithTimeout(context.TODO(), 300*time.Millisecond)
		defer cancel()

		listenerImpl.Listen(ctx)
	})

	t.Run("success: callback", func(t *testing.T) {
		t.Parallel()

		botAPIMock := mocks.BotAPIMock{
			GetUpdatesChanFunc: func(config tbapi.UpdateConfig) tbapi.UpdatesChannel {
				cfg := tbapi.NewUpdate(0)
				cfg.Timeout = 60
				assert.Equal(t, cfg, config)

				ch := make(chan tbapi.Update, 1)
				ch <- tbapi.Update{
					CallbackQuery: &tbapi.CallbackQuery{
						From: &tbapi.User{
							ID:       2,
							UserName: "Nirav Martini",
						},
						Data: "winds",
						Message: &tbapi.Message{
							MessageID: 13246,
							Chat: &tbapi.Chat{
								ID: 1,
							},
						},
					},
				}
				return ch
			},
		}
		updateReceiverMock := mocks.UpdateReceiverMock{
			OnCallbackQueryFunc: func(_ context.Context, callback domain.TgCallbackQuery) error {
				assert.Equal(t, domain.TgCallbackQuery{
					ChatID:   1,
					UserID:   2,
					UserName: "Nirav Martini",
					Data:     "winds",
				}, callback)
				return nil
			},
		}

		listenerImpl := New(&botAPIMock, &updateReceiverMock)

		ctx, cancel := context.WithTimeout(context.TODO(), 300*time.Millisecond)
		defer cancel()

		listenerImpl.Listen(ctx)
	})

	t.Run("updates chan is closed", func(t *testing.T) {
		t.Parallel()

		botAPIMock := mocks.BotAPIMock{
			GetUpdatesChanFunc: func(config tbapi.UpdateConfig) tbapi.UpdatesChannel {
				cfg := tbapi.NewUpdate(0)
				cfg.Timeout = 60
				assert.Equal(t, cfg, config)

				ch := make(chan tbapi.Update, 1)
				close(ch)
				return ch
			},
		}

		listenerImpl := New(&botAPIMock, nil)

		ctx, cancel := context.WithTimeout(context.TODO(), 300*time.Millisecond)
		defer cancel()

		listenerImpl.Listen(ctx)
	})

	t.Run("can't process update", func(t *testing.T) {
		t.Parallel()

		botAPIMock := mocks.BotAPIMock{
			GetUpdatesChanFunc: func(config tbapi.UpdateConfig) tbapi.UpdatesChannel {
				cfg := tbapi.NewUpdate(0)
				cfg.Timeout = 60
				assert.Equal(t, cfg, config)

				ch := make(chan tbapi.Update, 1)
				ch <- tbapi.Update{
					Message: &tbapi.Message{
						MessageID: 13246,
						Text:      "winds",
						From: &tbapi.User{
							ID:       2,
							UserName: "Nirav Martini",
						},
						Chat: &tbapi.Chat{
							ID: 1,
						},
					},
				}
				return ch
			},
		}
		updateReceiverMock := mocks.UpdateReceiverMock{
			OnMessageFunc: func(_ context.Context, message domain.TgMessage) error {
				return errors.New("can't process update")
			},
		}

		listenerImpl := New(&botAPIMock, &updateReceiverMock)

		ctx, cancel := context.WithTimeout(context.TODO(), 300*time.Millisecond)
		defer cancel()

		listenerImpl.Listen(ctx)
	})

}

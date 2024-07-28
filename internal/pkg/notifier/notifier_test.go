package notifier

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/notifier/mocks"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
	"github.com/stretchr/testify/assert"
)

func TestNotifier_Run(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		const (
			reminderID int64 = 436745
			userID     int64 = 4357
			chatID     int64 = 4568
		)

		senderMock := mocks.BotResponseSenderMock{
			SendBotResponseFunc: func(response sender.BotResponse, opts ...sender.BotResponseOption) error {
				a := assert.New(t)

				a.Equal(sender.BotResponse{
					ChatID: chatID,
					Text:   "FooBar",
				}, response)

				a.Len(opts, 1)
				return nil
			},
		}
		storageMock := mocks.StorageMock{
			GetPendingRemindersFunc: func(ctx context.Context, limit int64) ([]domain.Reminder, error) {
				return []domain.Reminder{
					{
						ID:           reminderID,
						ChatID:       chatID,
						UserID:       userID,
						Text:         "FooBar",
						Status:       domain.ReminderStatusPending,
						AttemptsLeft: 3,
					},
				}, nil
			},
			UpdateReminderFunc: func(ctx context.Context, reminder domain.Reminder) error {
				assert.Equal(t, reminderID, reminder.ID)
				assert.Equal(t, chatID, reminder.ChatID)
				assert.Equal(t, userID, reminder.UserID)
				assert.EqualValues(t, 2, reminder.AttemptsLeft)
				assert.Equal(t, domain.ReminderStatusPending, reminder.Status)
				assert.WithinDuration(t, timeNowUTC().Add(15*time.Minute), reminder.RemindAt, 1*time.Second)
				return nil
			},
		}

		notifierImpl := New(&senderMock, &storageMock, 300*time.Millisecond)

		ctx, cancel := context.WithTimeout(context.TODO(), 400*time.Millisecond)
		defer cancel()

		notifierImpl.Run(ctx)
	})

	t.Run("error: context canceled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.TODO(), 400*time.Millisecond)
		cancel()

		notifierImpl := New(nil, nil, 300*time.Millisecond)
		notifierImpl.Run(ctx)
	})

	t.Run("error: can't get pending reminders", func(t *testing.T) {
		t.Parallel()

		const (
			reminderID int64 = 436745
			userID     int64 = 4357
			chatID     int64 = 4568
		)

		storageMock := mocks.StorageMock{
			GetPendingRemindersFunc: func(ctx context.Context, limit int64) ([]domain.Reminder, error) {
				return nil, errors.New("some error")
			},
		}

		notifierImpl := New(nil, &storageMock, 300*time.Millisecond)

		ctx, cancel := context.WithTimeout(context.TODO(), 400*time.Millisecond)
		defer cancel()

		notifierImpl.Run(ctx)
	})

	t.Run("error: can't send bot response", func(t *testing.T) {
		t.Parallel()

		const (
			reminderID int64 = 436745
			userID     int64 = 4357
			chatID     int64 = 4568
		)

		senderMock := mocks.BotResponseSenderMock{
			SendBotResponseFunc: func(response sender.BotResponse, opts ...sender.BotResponseOption) error {
				return errors.New("some error")
			},
		}
		storageMock := mocks.StorageMock{
			GetPendingRemindersFunc: func(ctx context.Context, limit int64) ([]domain.Reminder, error) {
				return []domain.Reminder{
					{
						ID:           reminderID,
						ChatID:       chatID,
						UserID:       userID,
						Text:         "FooBar",
						Status:       domain.ReminderStatusPending,
						AttemptsLeft: 3,
					},
				}, nil
			},
			UpdateReminderFunc: func(ctx context.Context, reminder domain.Reminder) error {
				assert.Equal(t, reminderID, reminder.ID)
				assert.Equal(t, chatID, reminder.ChatID)
				assert.Equal(t, userID, reminder.UserID)
				assert.EqualValues(t, 2, reminder.AttemptsLeft)
				assert.Equal(t, domain.ReminderStatusPending, reminder.Status)
				assert.WithinDuration(t, timeNowUTC().Add(15*time.Minute), reminder.RemindAt, 1*time.Second)
				return nil
			},
		}

		notifierImpl := New(&senderMock, &storageMock, 300*time.Millisecond)

		ctx, cancel := context.WithTimeout(context.TODO(), 400*time.Millisecond)
		defer cancel()

		notifierImpl.Run(ctx)
	})

	t.Run("error: can't update reminder", func(t *testing.T) {
		t.Parallel()

		const (
			reminderID int64 = 436745
			userID     int64 = 4357
			chatID     int64 = 4568
		)

		senderMock := mocks.BotResponseSenderMock{
			SendBotResponseFunc: func(response sender.BotResponse, opts ...sender.BotResponseOption) error {
				a := assert.New(t)

				a.Equal(sender.BotResponse{
					ChatID: chatID,
					Text:   "FooBar",
				}, response)

				a.Len(opts, 1)
				return nil
			},
		}
		storageMock := mocks.StorageMock{
			GetPendingRemindersFunc: func(ctx context.Context, limit int64) ([]domain.Reminder, error) {
				return []domain.Reminder{
					{
						ID:           reminderID,
						ChatID:       chatID,
						UserID:       userID,
						Text:         "FooBar",
						Status:       domain.ReminderStatusPending,
						AttemptsLeft: 3,
					},
				}, nil
			},
			UpdateReminderFunc: func(ctx context.Context, reminder domain.Reminder) error {
				return errors.New("some error")
			},
		}

		notifierImpl := New(&senderMock, &storageMock, 300*time.Millisecond)

		ctx, cancel := context.WithTimeout(context.TODO(), 400*time.Millisecond)
		defer cancel()

		notifierImpl.Run(ctx)
	})

	t.Run("success: attempts exhausted", func(t *testing.T) {
		t.Parallel()

		const (
			reminderID int64 = 436745
			userID     int64 = 4357
			chatID     int64 = 4568
		)

		senderMock := mocks.BotResponseSenderMock{
			SendBotResponseFunc: func(response sender.BotResponse, opts ...sender.BotResponseOption) error {
				a := assert.New(t)

				a.Equal(sender.BotResponse{
					ChatID: chatID,
					Text:   "FooBar",
				}, response)

				a.Len(opts, 1)
				return nil
			},
		}
		storageMock := mocks.StorageMock{
			GetPendingRemindersFunc: func(ctx context.Context, limit int64) ([]domain.Reminder, error) {
				return []domain.Reminder{
					{
						ID:           reminderID,
						ChatID:       chatID,
						UserID:       userID,
						Text:         "FooBar",
						Status:       domain.ReminderStatusPending,
						AttemptsLeft: 1,
					},
				}, nil
			},
			UpdateReminderFunc: func(ctx context.Context, reminder domain.Reminder) error {
				assert.Equal(t, reminderID, reminder.ID)
				assert.Equal(t, chatID, reminder.ChatID)
				assert.Equal(t, userID, reminder.UserID)
				assert.Zero(t, reminder.AttemptsLeft)
				assert.Equal(t, domain.ReminderStatusAttemptsExhausted, reminder.Status)
				assert.WithinDuration(t, timeNowUTC().Add(15*time.Minute), reminder.RemindAt, 1*time.Second)
				return nil
			},
		}

		notifierImpl := New(&senderMock, &storageMock, 300*time.Millisecond)

		ctx, cancel := context.WithTimeout(context.TODO(), 400*time.Millisecond)
		defer cancel()

		notifierImpl.Run(ctx)
	})

}

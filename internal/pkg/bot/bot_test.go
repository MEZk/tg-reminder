package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mezk/tg-reminder/internal/pkg/bot/mocks"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
	"github.com/mezk/tg-reminder/internal/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestBot_OnCallbackQuery(t *testing.T) {
	const (
		expChatID   int64 = 43548
		expUserID   int64 = 546567
		expUserName       = "johndoe"
	)

	var dbError = errors.New("db error")

	testCases := []struct {
		name     string
		message  domain.TgCallbackQuery
		now      time.Time
		setMocks func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock)
		expErr   string
	}{
		// success
		{
			name: "success: done reminder button",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_reminder_done/12345",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, botState)
					return nil
				}
				store.SetReminderStatusFunc = func(_ context.Context, id int64, status domain.ReminderStatus) error {
					a.EqualValues(12345, id)
					a.Equal(domain.ReminderStatusDone, status)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Я пометил напоминание как выполненное!",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: remove reminder button",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_remove_reminder",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameRemoveReminder,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Напиши номер напоминания для удаления.",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: delay reminder button",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_delay_reminder/12345/1h",
			},
			now: time.Date(2024, 1, 1, 11, 1, 1, 0, time.UTC),
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, botState)
					return nil
				}
				store.DelayReminderFunc = func(_ context.Context, id int64, remindAt time.Time) error {
					a.EqualValues(12345, id)
					a.Equal(time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC), remindAt)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.EqualValues(expChatID, response.ChatID)
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Я отложил напоминание. Напомню позже *2024-01-01 15:01*!",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: remind at button",
			now:  time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC),
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_remind_at/time/20:30",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					a.EqualValues(expUserID, userID)
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameEnterReminAt,
						Context: &domain.BotStateContext{
							ReminderText: "FooBarBaz",
						},
					}, nil
				}
				store.SaveReminderFunc = func(_ context.Context, reminder domain.Reminder) (int64, error) {
					a.Equal(domain.Reminder{
						ChatID:       expChatID,
						UserID:       expUserID,
						Text:         "FooBarBaz",
						RemindAt:     time.Date(2024, 1, 1, 17, 30, 0, 0, time.UTC),
						Status:       domain.ReminderStatusPending,
						AttemptsLeft: 3,
					}, reminder)
					return 1, nil
				}
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "*2024-01-01 20:30* я напомню тебе о *FooBarBaz*!",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: edit reminder button",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_edit_reminder",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameEditReminder,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Напиши номер напоминания для редактирования.",
					}, response)
					return nil
				}
			},
		},

		// error
		{
			name: "error: done reminder button, can't set reminder status",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_reminder_done/12345",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return nil
				}
				store.SetReminderStatusFunc = func(_ context.Context, id int64, status domain.ReminderStatus) error {
					return dbError
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error: done reminder button, can't get reminder id",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_reminder_done/foobarbaz",
			},
			expErr: `can'p parse reminderID: strconv.ParseInt: parsing "foobarbaz": invalid syntax`,
		},
		{
			name: "error: done reminder button, can't save bot status",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_reminder_done/12345",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return dbError
				}
				store.SetReminderStatusFunc = func(_ context.Context, id int64, status domain.ReminderStatus) error {
					return nil
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error: remove reminder button, can't save bot state",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_remove_reminder",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return dbError
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error: delay reminder button, can't parse reminder id",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_delay_reminder/foobar/1h",
			},
			expErr: `can't parse reminderID: strconv.ParseInt: parsing "foobar": invalid syntax`,
		},
		{
			name: "success: delay reminder button",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_delay_reminder/12345/fooa1h",
			},
			expErr: `can't parse delay: failed to parse remindAt duration: time: invalid duration "fooa1h"`,
		},
		{
			name: "success: delay reminder button, can't delay reminder",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_delay_reminder/12345/1h",
			},
			now: time.Date(2024, 1, 1, 11, 1, 1, 0, time.UTC),
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.DelayReminderFunc = func(_ context.Context, id int64, remindAt time.Time) error {
					return dbError
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "success: delay reminder button, can't save bot state",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_delay_reminder/12345/1h",
			},
			now: time.Date(2024, 1, 1, 11, 1, 1, 0, time.UTC),
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return dbError
				}
				store.DelayReminderFunc = func(_ context.Context, id int64, remindAt time.Time) error {
					return nil
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error: remind at button",
			now:  time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC),
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_remind_at/time/foo20:30",
			},
			expErr: `failed to parse remindAt time: parsing time "foo20:30" as "15:04": cannot parse "foo20:30" as "15"`,
		},
		{
			name: "error: edit reminder button, can't save bot state",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_edit_reminder",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return dbError
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error: unknown button",
			message: domain.TgCallbackQuery{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Data:     "btn_foo",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Я не понимаю о чём речь! Пожалуйста, воспользуйся командой /help.",
					}, response)
					return nil
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if !tc.now.IsZero() {
				tmpTimeNowUTC := timeNowUTC
				defer func() {
					timeNowUTC = tmpTimeNowUTC
				}()
				timeNowUTC = func() time.Time {
					return tc.now
				}
			}

			a := assert.New(t)
			senderMock := &mocks.ResponseSenderMock{}
			storeMock := &mocks.StorageMock{}
			if tc.setMocks != nil {
				tc.setMocks(a, senderMock, storeMock)
			}

			botImpl := New(senderMock, storeMock)

			actErr := botImpl.OnCallbackQuery(context.TODO(), tc.message)

			if tc.expErr != "" {
				a.EqualError(actErr, tc.expErr)
			} else {
				a.NoError(actErr)
			}
		})
	}
}

func TestBot_OnMessage(t *testing.T) {
	const (
		expChatID   int64 = 43548
		expUserID   int64 = 546567
		expUserName       = "johndoe"
	)

	var dbError = errors.New("db error")

	testCases := []struct {
		name     string
		message  domain.TgMessage
		now      time.Time
		setMocks func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock)
		expErr   string
	}{
		// commands
		{
			name: "success: start cmd",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/start",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveUserFunc = func(_ context.Context, user domain.User) error {
					a.Equal(domain.User{
						ID:     expUserID,
						Name:   expUserName,
						Status: domain.UserStatusActive,
					}, user)
					return nil
				}
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Привет, @johndoe!\nТеперь ты можешь со мной работать.\nДля справки используй команду /help.",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: help cmd",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/help",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameHelp,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "\n\t\tСписок доступных команд:\n\t\t- /start - cправка\n\t\t- /help - начать работу с ботом\n\t\t- /create\\_reminder - cоздать напоминание\n\t\t- /enable\\_reminders - включить напоминания\n\t\t- /disable\\_reminders - выключить напоминания\n\t\t- /my\\_reminders - мои напоминания",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: create reminder cmd",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/create_reminder",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameCreateReminder,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "О чём напомнить?",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: my reminders cmd",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/my_reminders",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetMyRemindersFunc = func(_ context.Context, userID int64, chatID int64) ([]domain.Reminder, error) {
					a.Equal(expUserID, userID)
					a.Equal(expChatID, chatID)
					return []domain.Reminder{
						{
							ID:       12,
							ChatID:   expChatID,
							UserID:   expUserID,
							Text:     "Напоминание 1",
							RemindAt: time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC),
						},
					}, nil
				}

				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameMyReminders,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "*Вот список твоих активных напоминаний:*\n12. 2024-01-01 04:01: Напоминание 1",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: enable reminders cmd",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/enable_reminders",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameEnableReminders,
					}, botState)
					return nil
				}
				store.SetUserStatusFunc = func(_ context.Context, id int64, status domain.UserStatus) error {
					a.Equal(expUserID, id)
					a.Equal(domain.UserStatusActive, status)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Уведомления включены. Для отключения уведомлений воспользуйтесь командой /disable\\_reminders.",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: disable reminders cmd",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/disable_reminders",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameDisableReminders,
					}, botState)
					return nil
				}
				store.SetUserStatusFunc = func(_ context.Context, id int64, status domain.UserStatus) error {
					a.Equal(expUserID, id)
					a.Equal(domain.UserStatusInactive, status)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Уведомления отключены. Для включения уведомлений воспользуйтесь командой /enable\\_reminders.",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: my reminders cmd, no reminders found",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/my_reminders",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetMyRemindersFunc = func(_ context.Context, userID int64, chatID int64) ([]domain.Reminder, error) {
					return nil, nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "У тебя нет напоминаний!",
					}, response)
					return nil
				}
			},
		},

		// messages from user
		{
			name: "success: msg with reminder text",
			now:  time.Date(2024, 1, 1, 1, 1, 0, 0, time.UTC),
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "Punishment lawyer blank arrives luis deviant failing, grocery feb.",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					a.Equal(expUserID, userID)
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameCreateReminder,
					}, nil
				}

				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID:  expUserID,
						Name:    domain.BotStateNameEnterReminAt,
						Context: &domain.BotStateContext{ReminderText: "Punishment lawyer blank arrives luis deviant failing, grocery feb."},
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, opts ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "*Когда напомнить?*\n\nТекущая дата и время (Москва): 2024-01-01 04:01.\n\nВведи дату в формате *YYYY-MM-DD HH:mm*.\n\t\nНапример, 2024-06-07 11:30 значит, что я пришлю тебе напоминание 7 мая 2024 года в 11:30.\n\nИли выберите опцию ниже:",
					}, response)
					a.Len(opts, 1)
					return nil
				}
			},
		},
		{
			name: "success: msg with remind_at",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "2024-01-01 04:01",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameEnterReminAt,
						Context: &domain.BotStateContext{
							ReminderText: "FooBarBaz",
						},
					}, nil
				}

				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, botState)
					return nil
				}

				store.SaveReminderFunc = func(_ context.Context, reminder domain.Reminder) (int64, error) {
					a.Equal(domain.Reminder{
						ChatID:       expChatID,
						UserID:       expUserID,
						Text:         "FooBarBaz",
						RemindAt:     time.Date(2024, 1, 1, 1, 1, 0, 0, time.UTC),
						Status:       domain.ReminderStatusPending,
						AttemptsLeft: 3,
					}, reminder)
					return 1, nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   `*2024-01-01 04:01* я напомню тебе о *FooBarBaz*!`,
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: msg with reminder id to remove",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "12345",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameRemoveReminder,
					}, nil
				}

				store.RemoveReminderFunc = func(ctx context.Context, id int64) error {
					a.EqualValues(12345, id)
					return nil
				}

				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Напоминание 12345 удалено!",
					}, response)
					return nil
				}
			},
		},
		{
			name: "success: unsupported response",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/foo-bar-baz",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					a.Equal(expUserID, userID)
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Я не понимаю о чём речь! Пожалуйста, воспользуйся командой /help.",
					}, response)
					return nil
				}
			},
		},

		// error cases
		{
			name: "error: start cmd, user already exists",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/start",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveUserFunc = func(_ context.Context, user domain.User) error {
					return storage.ErrUserAlreadyExists
				}
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "@johndoe, ранее мы уже начали общение, предлагаю продолжить!",
					}, response)
					return nil
				}
			},
		},
		{
			name: "error: start cmd, save user db error",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/start",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveUserFunc = func(_ context.Context, user domain.User) error {
					return errors.New("db error")
				}
			},
			expErr: `db error`,
		},
		{
			name: "error: start cmd, user already exists, save bot state error",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/start",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveUserFunc = func(_ context.Context, user domain.User) error {
					return storage.ErrUserAlreadyExists
				}
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return errors.New("db error")
				}
			},
			expErr: `db error`,
		},
		{
			name: "error: help cmd, save bot state error",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/help",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return errors.New("db error")
				}
			},
			expErr: `db error`,
		},
		{
			name: "error: create reminder cmd, save bot state error",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/create_reminder",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return errors.New("db error")
				}
			},
			expErr: `db error`,
		},
		{
			name: "error: my reminders cmd, get reminders error",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/my_reminders",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetMyRemindersFunc = func(_ context.Context, userID int64, chatID int64) ([]domain.Reminder, error) {
					return nil, errors.New("db error")
				}
			},
			expErr: `db error`,
		},
		{
			name: "errors: enable reminders cmd, set user status error",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "/enable_reminders",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.SetUserStatusFunc = func(_ context.Context, id int64, status domain.UserStatus) error {
					return errors.New("db error")
				}
			},
			expErr: `db error`,
		},
		{
			name: "error: msg with reminder text, can't save bot state",
			now:  time.Date(2024, 1, 1, 1, 1, 0, 0, time.UTC),
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "Punishment lawyer blank arrives luis deviant failing, grocery feb.",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					a.Equal(expUserID, userID)
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameCreateReminder,
					}, nil
				}

				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return dbError
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error: msg with remind_at, can't get remind at",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "foo",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameEnterReminAt,
						Context: &domain.BotStateContext{
							ReminderText: "FooBarBaz",
						},
					}, nil
				}
			},
			expErr: `can't parse remindAt foo: parsing time "foo" as "2006-01-02 15:04": cannot parse "foo" as "2006"`,
		},
		{
			name: "error: msg with reminder id to remove, invalid id",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "foo",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameRemoveReminder,
					}, nil
				}
			},
			expErr: `failed to parse reminder id foo: strconv.ParseInt: parsing "foo": invalid syntax`,
		},
		{
			name: "error: msg with reminder id to remove, reminder is not found",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "12345",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameRemoveReminder,
					}, nil
				}

				store.RemoveReminderFunc = func(ctx context.Context, id int64) error {
					return storage.ErrReminderNotFound
				}

				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					a.Equal(domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameStart,
					}, botState)
					return nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, _ ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Напоминание 12345 не найдено!",
					}, response)
					return nil
				}
			},
		},
		{
			name: "error: msg with reminder id to remove, db error",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "12345",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameRemoveReminder,
					}, nil
				}

				store.RemoveReminderFunc = func(ctx context.Context, id int64) error {
					return dbError
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error: msg with reminder id to remove, can't save bot state",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "12345",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameRemoveReminder,
					}, nil
				}

				store.RemoveReminderFunc = func(ctx context.Context, id int64) error {
					a.EqualValues(12345, id)
					return nil
				}

				store.SaveBotStateFunc = func(_ context.Context, botState domain.BotState) error {
					return dbError
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error: can't get bot state",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "Punishment lawyer blank arrives luis deviant failing, grocery feb.",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{}, dbError
				}
			},
			expErr: dbError.Error(),
		},
		{
			name: "error:unsupported",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "Punishment lawyer blank arrives luis deviant failing, grocery feb.",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   "foobar",
					}, nil
				}

				responseSender.SendBotResponseFunc = func(response sender.BotResponse, opts ...sender.BotResponseOption) error {
					a.Equal(sender.BotResponse{
						ChatID: expChatID,
						Text:   "Я не понимаю о чём речь! Пожалуйста, воспользуйся командой /help.",
					}, response)
					return nil
				}
			},
		},
		{
			name: "errorr: edit reminder, unimplemented",
			message: domain.TgMessage{
				ChatID:   expChatID,
				UserID:   expUserID,
				UserName: expUserName,
				Text:     "foo",
			},
			setMocks: func(a *assert.Assertions, responseSender *mocks.ResponseSenderMock, store *mocks.StorageMock) {
				store.GetBotStateFunc = func(_ context.Context, userID int64) (domain.BotState, error) {
					return domain.BotState{
						UserID: expUserID,
						Name:   domain.BotStateNameEditReminder,
					}, nil
				}
			},
			expErr: errors.ErrUnsupported.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if !tc.now.IsZero() {
				tmpTimeNowUTC := timeNowUTC
				defer func() {
					timeNowUTC = tmpTimeNowUTC
				}()
				timeNowUTC = func() time.Time {
					return tc.now
				}
			}

			a := assert.New(t)
			senderMock := &mocks.ResponseSenderMock{}
			storeMock := &mocks.StorageMock{}
			if tc.setMocks != nil {
				tc.setMocks(a, senderMock, storeMock)
			}

			botImpl := New(senderMock, storeMock)

			actErr := botImpl.OnMessage(context.TODO(), tc.message)

			if tc.expErr != "" {
				a.EqualError(actErr, tc.expErr)
			} else {
				a.NoError(actErr)
			}
		})
	}
}

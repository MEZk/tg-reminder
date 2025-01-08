package sender

import (
	"errors"
	"testing"

	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func Test_botResponseSender_SendBotResponse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		resp     BotResponse
		opts     []BotResponseOption
		setMocks func(r *assert.Assertions, botAPIMock *BotAPIMock)
		expErr   string
	}{
		{
			name: "success: no options",
			resp: BotResponse{
				ChatID:           2,
				ReplyToMessageID: 4,
				Text:             "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
			},
			setMocks: func(a *assert.Assertions, botAPIMock *BotAPIMock) {
				botAPIMock.SendFunc = func(c tbapi.Chattable) (tbapi.Message, error) {
					a.Equal(tbapi.MessageConfig{
						BaseChat: tbapi.BaseChat{
							ChatID:           2,
							ReplyToMessageID: 4,
						},
						Text:                  "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
						ParseMode:             "Markdown",
						DisableWebPagePreview: true,
					}, c)
					return tbapi.Message{}, nil
				}
			},
		},
		{
			name: "success: send as plain text",
			resp: BotResponse{
				ChatID:           2,
				ReplyToMessageID: 4,
				Text:             "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
			},
			setMocks: func(a *assert.Assertions, botAPIMock *BotAPIMock) {
				firstCall := true
				botAPIMock.SendFunc = func(c tbapi.Chattable) (tbapi.Message, error) {
					if firstCall {
						firstCall = false
						return tbapi.Message{}, errors.New("some error")
					}

					a.Equal(tbapi.MessageConfig{
						BaseChat: tbapi.BaseChat{
							ChatID:           2,
							ReplyToMessageID: 4,
						},
						Text:                  "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
						ParseMode:             "",
						DisableWebPagePreview: true,
					}, c)
					return tbapi.Message{}, nil
				}
			},
		},
		{
			name: "error: telegram api returns error",
			resp: BotResponse{
				ChatID:           2,
				ReplyToMessageID: 4,
				Text:             "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
			},
			setMocks: func(a *assert.Assertions, botAPIMock *BotAPIMock) {
				botAPIMock.SendFunc = func(c tbapi.Chattable) (tbapi.Message, error) {
					return tbapi.Message{}, errors.New("some internal error")
				}
			},
			expErr: `can't send message to telegram "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.": some internal error`,
		},
		{
			name: "success: WithMyRemindersListEditButtons option",
			resp: BotResponse{
				ChatID:           2,
				ReplyToMessageID: 4,
				Text:             "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
			},
			opts: []BotResponseOption{WithMyRemindersListEditButtons()},
			setMocks: func(a *assert.Assertions, botAPIMock *BotAPIMock) {
				botAPIMock.SendFunc = func(c tbapi.Chattable) (tbapi.Message, error) {
					a.Equal(tbapi.MessageConfig{
						BaseChat: tbapi.BaseChat{
							ChatID:           2,
							ReplyToMessageID: 4,
							ReplyMarkup: tbapi.NewInlineKeyboardMarkup(
								tbapi.NewInlineKeyboardRow(
									tbapi.NewInlineKeyboardButtonData("📝 Редактировать", "btn_edit_reminder"),
									tbapi.NewInlineKeyboardButtonData("❌ Удалить", "btn_remove_reminder"),
								),
							),
						},
						Text:                  "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
						ParseMode:             "Markdown",
						DisableWebPagePreview: true,
					}, c)
					return tbapi.Message{}, nil
				}
			},
		},
		{
			name: "success: WithReminderDatesButtons option",
			resp: BotResponse{
				ChatID:           2,
				ReplyToMessageID: 4,
				Text:             "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
			},
			opts: []BotResponseOption{WithReminderDatesButtons()},
			setMocks: func(a *assert.Assertions, botAPIMock *BotAPIMock) {
				botAPIMock.SendFunc = func(c tbapi.Chattable) (tbapi.Message, error) {
					a.Equal(tbapi.MessageConfig{
						BaseChat: tbapi.BaseChat{
							ChatID:           2,
							ReplyToMessageID: 4,
							ReplyMarkup: tbapi.NewInlineKeyboardMarkup(
								tbapi.NewInlineKeyboardRow(
									tbapi.NewInlineKeyboardButtonData("11:30", "btn_remind_at/time/11:30"),
									tbapi.NewInlineKeyboardButtonData("14:30", "btn_remind_at/time/14:30"),
									tbapi.NewInlineKeyboardButtonData("19:30", "btn_remind_at/time/19:30"),
									tbapi.NewInlineKeyboardButtonData("20:30", "btn_remind_at/time/20:30"),
								),
								tbapi.NewInlineKeyboardRow(
									tbapi.NewInlineKeyboardButtonData("30 мин", "btn_remind_at/duration/30m"),
									tbapi.NewInlineKeyboardButtonData("80 мин", "btn_remind_at/duration/80m"),
									tbapi.NewInlineKeyboardButtonData("1 день", "btn_remind_at/duration/24h"),
									tbapi.NewInlineKeyboardButtonData("1 месяц", "btn_remind_at/duration/730h"),
								),
							),
						},
						Text:                  "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
						ParseMode:             "Markdown",
						DisableWebPagePreview: true,
					}, c)
					return tbapi.Message{}, nil
				}
			},
		},
		{
			name: "success: WithReminderDoneButton option",
			resp: BotResponse{
				ChatID:           2,
				ReplyToMessageID: 4,
				Text:             "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
			},
			opts: []BotResponseOption{WithReminderDoneButton(12345)},
			setMocks: func(a *assert.Assertions, botAPIMock *BotAPIMock) {
				botAPIMock.SendFunc = func(c tbapi.Chattable) (tbapi.Message, error) {
					a.Equal(tbapi.MessageConfig{
						BaseChat: tbapi.BaseChat{
							ChatID:           2,
							ReplyToMessageID: 4,
							ReplyMarkup: tbapi.NewInlineKeyboardMarkup(
								tbapi.NewInlineKeyboardRow(
									tbapi.NewInlineKeyboardButtonData("🔄 30 мин.", "btn_delay_reminder/12345/30m"),
									tbapi.NewInlineKeyboardButtonData("🔄 80 мин.", "btn_delay_reminder/12345/80m"),
									tbapi.NewInlineKeyboardButtonData("🔄 3 час.", "btn_delay_reminder/12345/3h"),
								),
								tbapi.NewInlineKeyboardRow(
									tbapi.NewInlineKeyboardButtonData("🔄 1 ден.", "btn_delay_reminder/12345/24h"),
									tbapi.NewInlineKeyboardButtonData("🔄 1 нед.", "btn_delay_reminder/12345/168h"),
									tbapi.NewInlineKeyboardButtonData("🔄 1 мес.", "btn_delay_reminder/12345/730h"),
								),
								tbapi.NewInlineKeyboardRow(
									tbapi.NewInlineKeyboardButtonData("✅ Готово", "btn_reminder_done/12345"),
								),
							),
						},
						Text:                  "Pipeline arts speakers realized choose aviation thong, adopt events switching info platforms units specialized, particular pants compatibility determines attachments pee assignment, licking tradition fool synthetic survivors denial alice.",
						ParseMode:             "Markdown",
						DisableWebPagePreview: true,
					}, c)
					return tbapi.Message{}, nil
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// ARRANGE
			a := assert.New(t)
			botAPIMock := BotAPIMock{}

			if tc.setMocks != nil {
				tc.setMocks(a, &botAPIMock)
			}

			senderImpl := New(&botAPIMock)

			// ACT
			err := senderImpl.SendBotResponse(tc.resp, tc.opts...)

			// ASSERT
			if tc.expErr != "" {
				a.EqualError(err, tc.expErr)
			} else {
				a.NoError(err)
			}
		})
	}
}

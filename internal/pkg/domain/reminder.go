package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DefaultAttemptsLeft - default attempts left to deliver a reminder.
const DefaultAttemptsLeft = 10

// Reminder - reminder representation.
type Reminder struct {
	ID           int64          `db:"id"`
	ChatID       int64          `db:"chat_id"`
	UserID       int64          `db:"user_id"`
	Text         string         `db:"text"`
	CreatedAt    time.Time      `db:"created_at"`
	ModifiedAt   time.Time      `db:"modified_at"`
	RemindAt     time.Time      `db:"remind_at"`
	Status       ReminderStatus `db:"status"`
	AttemptsLeft byte           `db:"attempts_left"`
}

func (r Reminder) String() string {
	return fmt.Sprintf("[ID: %d, UserID: %d, ChatID: %d, Status: %s, RemindAt: %s, AttemptsLeft: %d, Data: %s]", r.ID, r.UserID, r.ChatID, r.Status, r.RemindAt, r.AttemptsLeft, r.Text)
}

const layoutTimeOnly = "15:04"

// FormatList - format reminder info to send to user as an entity of reminders list.
func (r Reminder) FormatList(now time.Time) string {
	var (
		remindAtMSK         = MoscowTime(r.RemindAt)
		nYear, nMonth, nDay = MoscowTime(now).Date()
		rYear, rMonth, rDay = remindAtMSK.Date()
		timeOnly            = remindAtMSK.Format(layoutTimeOnly)
	)

	var sb strings.Builder
	sb.WriteString(EmojiWhiteHeavyCheckMark)
	sb.WriteString(" *")
	sb.WriteString(r.Text)
	sb.WriteString("*")

	if nYear == rYear && nMonth == rMonth && nDay == rDay { // today
		sb.WriteString(EmojiExclamationMark)
		sb.WriteString("\n")
		sb.WriteString(EmojiAlarmClock)
		sb.WriteString(" Сегодня ")
		sb.WriteString(timeOnly)
	} else {
		sb.WriteString("\n")
		sb.WriteString(EmojiAlarmClock)
		sb.WriteRune(' ')
		sb.WriteString(strconv.Itoa(remindAtMSK.Day()))
		sb.WriteRune(' ')
		sb.WriteString(getRussianMonth(remindAtMSK.Month()))
		sb.WriteRune(' ')
		sb.WriteString(timeOnly)
	}

	sb.WriteString("\n")
	sb.WriteString(EmojiKeycapHash)
	sb.WriteRune(' ')
	sb.WriteString(strconv.FormatInt(r.ID, 10))

	return sb.String()
}

// FormatNotify - format reminder info to send to user as notification.
func (r Reminder) FormatNotify() string {
	return fmt.Sprintf("%[1]s*НАПОМИНАНИЕ*%[1]s\n\n*%[2]s*\n\nСегодня %[3]s%[4]s%[5]s\n\nЧтобы отложить напоминание используйте кнопки%[6]s%[7]s, расположенные ниже.",
		EmojiDoubleExclamationMark,
		strings.ToUpper(r.Text),
		MoscowTime(r.RemindAt).Format(layoutTimeOnly), NoBreakSpace, EmojiAlarmClock,
		NoBreakSpace, EmojiCounterclockwiseArrowsButton,
	)
}

// ReminderStatus - status of a remidner.
type ReminderStatus string

const (
	// ReminderStatusPending is a pending status. Reminder will be sent to user at remindAt time.
	ReminderStatusPending ReminderStatus = "pending"
	// ReminderStatusDone is a done status. Reminder was sent to user and user marked it as 'done'.
	ReminderStatusDone ReminderStatus = "done"
	// ReminderStatusAttemptsExhausted describes the situation in which all attempts to receive 'done' from user are finished.
	ReminderStatusAttemptsExhausted ReminderStatus = "attempts_exhausted"
)

func getRussianMonth(m time.Month) string {
	switch m {
	case time.January:
		return "янв."
	case time.February:
		return "фев."
	case time.March:
		return "мар."
	case time.April:
		return "апр."
	case time.May:
		return "мая"
	case time.June:
		return "июн."
	case time.July:
		return "июл."
	case time.August:
		return "авг."
	case time.September:
		return "сент."
	case time.October:
		return "окт."
	case time.November:
		return "нояб."
	case time.December:
		return "дек."
	default:
		return ""
	}
}

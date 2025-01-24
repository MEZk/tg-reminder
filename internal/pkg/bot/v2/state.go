package v2

import (
	"github.com/go-telegram/fsm"
)

const (
	stateDefault fsm.StateID = "default"

	stateCreateReminderAskText fsm.StateID = "create_reminder_ask_text"
	stateCreateReminderAskDate fsm.StateID = "create_reminder_ask_date"
	stateCreateReminderAskTime fsm.StateID = "create_reminder_ask_time"

	stateRemoveReminderAskID fsm.StateID = "remove_reminder_ask_id"

	stateEditReminderAskID   fsm.StateID = "edit_reminder_ask_id"
	stateEditReminderAskText fsm.StateID = "edit_reminder_ask_text"
	stateEditReminderAskDate fsm.StateID = "edit_reminder_ask_date"
	stateEditReminderAskTime fsm.StateID = "edit_reminder_ask_time"
)

type stateKey string

const (
	stateKeyReminderID       stateKey = "reminder_id"
	stateKeyReminderText     stateKey = "reminder_text"
	stateKeyReminderDateOnly stateKey = "reminder_date_only"
)

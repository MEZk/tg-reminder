package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBotState_ReminderID(t *testing.T) {
	t.Parallel()
	state := BotState{
		Context: &BotStateContext{ReminderID: 3126346},
	}
	assert.EqualValues(t, 3126346, state.ReminderID())

	state = BotState{Context: nil}
	assert.Zero(t, state.ReminderID())
}

func TestBotState_ReminderText(t *testing.T) {
	t.Parallel()
	state := BotState{
		Context: &BotStateContext{ReminderText: "stupid"},
	}
	assert.Equal(t, "stupid", state.ReminderText())

	var nilState *BotState
	assert.NotPanics(t, func() {
		nilState.SetReminderText("saf")
	})
}

func TestBotState_SetReminderID(t *testing.T) {
	t.Parallel()
	state := BotState{}
	state.SetReminderID(1366)
	assert.EqualValues(t, 1366, state.Context.ReminderID)

	var nilState *BotState
	assert.NotPanics(t, func() {
		nilState.SetReminderID(0)
	})
}

func TestBotState_SetReminderText(t *testing.T) {
	t.Parallel()
	state := BotState{}
	state.SetReminderText("FooBar")
	assert.Equal(t, "FooBar", state.Context.ReminderText)
}

func TestBotState_String(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "[UserID: 1, Name: Angelos Casados]", BotState{
		UserID: 1,
		Name:   "Angelos Casados",
	}.String())
}

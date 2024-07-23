package storage

import (
	"context"
	"time"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

func (s *storageTestSuite) Test_storage_GetBotState() {
	s.Run("success: bot state exists", func() {
		// ARRANGE
		state := domain.BotState{
			UserID:     1,
			Name:       domain.BotStateNameStart,
			ModifiedAt: time.Now().Truncate(1 * time.Minute),
			Context: &domain.BotStateContext{
				ReminderID:   1,
				ReminderText: "Went gang rogers teachers genome chargers directive, thing professionals.",
			},
		}

		s.NoError(s.storage.SaveBotState(context.TODO(), state))

		// ACT
		actState, err := s.storage.GetBotState(context.TODO(), state.UserID)

		// ASSERT
		s.NoError(err)
		s.Equal(state, actState)
	})

	s.Run("error: bot state is not found", func() {
		actState, err := s.storage.GetBotState(context.TODO(), 2)

		s.ErrorIs(err, ErrBotStateNotFound)
		s.Equal(domain.BotState{}, actState)
	})
}

func (s *storageTestSuite) Test_storage_SaveBotState() {
	s.Run("success: bot state does not exist", func() {
		// ARRANGE
		state := domain.BotState{
			UserID:     3,
			Name:       domain.BotStateNameStart,
			ModifiedAt: time.Now().Truncate(1 * time.Minute),
			Context: &domain.BotStateContext{
				ReminderID:   2,
				ReminderText: "Politics five verbal suited analyst minor worried, once jail improvement argument gloves dsl sussex, treasure neo contributions interaction fence.",
			},
		}

		// ACT
		s.NoError(s.storage.SaveBotState(context.TODO(), state))

		// ASSERT
		actState, err := s.storage.GetBotState(context.TODO(), state.UserID)
		s.NoError(err)
		s.Equal(state, actState)
	})

	s.Run("success: modified at is not set", func() {
		// ARRANGE
		state := domain.BotState{
			UserID: 3,
			Name:   domain.BotStateNameStart,
			Context: &domain.BotStateContext{
				ReminderID:   2,
				ReminderText: "Politics five verbal suited analyst minor worried, once jail improvement argument gloves dsl sussex, treasure neo contributions interaction fence.",
			},
		}

		// ACT
		s.NoError(s.storage.SaveBotState(context.TODO(), state))

		// ASSERT
		actState, err := s.storage.GetBotState(context.TODO(), state.UserID)
		s.NoError(err)
		s.NotZero(actState.ModifiedAt)
	})

	s.Run("error: bot state already exists", func() {
		// ARRANGE
		state := domain.BotState{
			UserID:     4,
			Name:       domain.BotStateNameStart,
			ModifiedAt: time.Now().Truncate(1 * time.Minute),
			Context: &domain.BotStateContext{
				ReminderID:   2,
				ReminderText: "Politics five verbal suited analyst minor worried, once jail improvement argument gloves dsl sussex, treasure neo contributions interaction fence.",
			},
		}

		// ACT
		s.NoError(s.storage.SaveBotState(context.TODO(), state))

		// ASSERT
		state.Name = domain.BotStateNameHelp
		state.Context.ReminderID = 2
		state.Context.ReminderText = "Bk endangered way fossil julie principle multimedia, routing rather ceramic netscape candy correctly pin, compatible carol concerts exercises ready circus pod, lesbians tourist provinces gray pace conducting patrol, properties."
		state.ModifiedAt = time.Now().Add(2 * time.Minute).Truncate(1 * time.Minute)

		s.NoError(s.storage.SaveBotState(context.TODO(), state), ErrBotStateAlreadyExists)

		actState, err := s.storage.GetBotState(context.TODO(), state.UserID)
		s.NoError(err)
		s.Equal(state, actState)
	})
}

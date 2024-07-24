package storage

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/stretchr/testify/require"
)

func (s *storageTestSuite) Test_storage_DelayReminder() {
	s.Run("success", func() {
		// ARRANGE
		reminder := domain.Reminder{
			ChatID:       1,
			UserID:       1,
			Text:         "Familiar indicated collection dense quest entirely undefined, engineers advertising curve Storage worker sony research. ",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}

		id, err := s.storage.SaveReminder(context.TODO(), reminder)
		s.NoError(err)

		// ACT
		remindAt := timeNowUTC().Truncate(1 * time.Minute)
		s.NoError(s.storage.DelayReminder(context.TODO(), id, remindAt))

		// ASSERT
		actReminder := s.mustGetReminder(id)
		s.EqualValues(3, actReminder.AttemptsLeft)
		s.Equal(remindAt, actReminder.RemindAt)
		s.Equal(domain.ReminderStatusPending, actReminder.Status)
		s.Greater(actReminder.ModifiedAt, reminder.ModifiedAt)
	})

	s.Run("error: not found", func() {
		s.ErrorIs(s.storage.DelayReminder(context.TODO(), 2513, timeNowUTC()), ErrReminderNotFound)
	})
}

func (s *storageTestSuite) Test_storage_GetMyReminders() {
	s.Run("success", func() {
		const (
			userID = 132436
			chatID = 3457547
		)

		pendingReminder := domain.Reminder{
			ChatID:       chatID,
			UserID:       userID,
			Text:         "Mechanisms fatal thought massage here lakes austria, qatar bless japanese consists bonds considerable hero.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}
		_, err := s.storage.SaveReminder(context.TODO(), pendingReminder)
		s.NoError(err)

		doneReminder := domain.Reminder{
			ChatID:       chatID,
			UserID:       userID,
			Text:         "Mechanisms fatal thought massage here lakes austria, qatar bless japanese consists bonds considerable hero.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusDone,
			AttemptsLeft: 3,
		}
		_, err = s.storage.SaveReminder(context.TODO(), doneReminder)
		s.NoError(err)

		actRemidners, err := s.storage.GetMyReminders(context.TODO(), userID, chatID)
		s.NoError(err)

		requireEqualRemindersList(s.Require(), []domain.Reminder{pendingReminder}, actRemidners)
	})
}

func (s *storageTestSuite) Test_storage_GetPendingRemidners() {
	s.Run("success: user is active", func() {
		const (
			userID = 464357
			chatID = 745674
		)

		pendingReminder1 := domain.Reminder{
			ChatID:       chatID,
			UserID:       userID,
			Text:         "Mechanisms fatal thought massage here lakes austria, qatar bless japanese consists bonds considerable hero.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}
		_, err := s.storage.SaveReminder(context.TODO(), pendingReminder1)
		s.NoError(err)

		pendingReminder2 := domain.Reminder{
			ChatID:       chatID,
			UserID:       userID,
			Text:         "Wisdom bankruptcy controls smart.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}
		_, err = s.storage.SaveReminder(context.TODO(), pendingReminder2)
		s.NoError(err)

		doneReminder := domain.Reminder{
			ChatID:       chatID,
			UserID:       userID,
			Text:         "Mechanisms fatal thought massage here lakes austria, qatar bless japanese consists bonds considerable hero.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusDone,
			AttemptsLeft: 3,
		}
		_, err = s.storage.SaveReminder(context.TODO(), doneReminder)
		s.NoError(err)

		s.NoError(s.storage.SaveUser(context.TODO(), domain.User{
			ID:     userID,
			Name:   "Danay Rodney",
			Status: domain.UserStatusActive,
		}))

		reminders, err := s.storage.GetPendingRemidners(context.TODO(), 2)
		s.NoError(err)
		requireEqualRemindersList(s.Require(), []domain.Reminder{pendingReminder1, pendingReminder2}, reminders)
	})

	s.Run("success: user is not active", func() {
		const (
			userID = 67854687
			chatID = 847686
		)

		pendingReminder1 := domain.Reminder{
			ChatID:       chatID,
			UserID:       userID,
			Text:         "Mechanisms fatal thought massage here lakes austria, qatar bless japanese consists bonds considerable hero.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}
		_, err := s.storage.SaveReminder(context.TODO(), pendingReminder1)
		s.NoError(err)

		s.NoError(s.storage.SaveUser(context.TODO(), domain.User{
			ID:     userID,
			Name:   "Danay Rodney",
			Status: domain.UserStatusInactive,
		}))

		reminders, err := s.storage.GetPendingRemidners(context.TODO(), 2)
		s.NoError(err)
		s.Empty(reminders)
	})

	s.Run("success: user is not found", func() {
		const (
			userID = 64563
			chatID = 364536
		)

		pendingReminder1 := domain.Reminder{
			ChatID:       chatID,
			UserID:       userID,
			Text:         "Mechanisms fatal thought massage here lakes austria, qatar bless japanese consists bonds considerable hero.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}
		_, err := s.storage.SaveReminder(context.TODO(), pendingReminder1)
		s.NoError(err)

		reminders, err := s.storage.GetPendingRemidners(context.TODO(), 2)
		s.NoError(err)
		s.Empty(reminders)
	})
}

func (s *storageTestSuite) Test_storage_RemoveReminder() {
	s.Run("success: reminder exists", func() {
		const (
			userID = 347657
			chatID = 7456724
		)

		pendingReminder := domain.Reminder{
			ChatID:       chatID,
			UserID:       userID,
			Text:         "Mechanisms fatal thought massage here lakes austria, qatar bless japanese consists bonds considerable hero.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}
		id, err := s.storage.SaveReminder(context.TODO(), pendingReminder)
		s.NoError(err)

		s.NoError(s.storage.RemoveReminder(context.TODO(), id))

		_, err = s.getGetReminder(id)
		s.ErrorIs(err, sql.ErrNoRows)
	})

	s.Run("success: reminder does not exist", func() {
		s.ErrorIs(s.storage.RemoveReminder(context.TODO(), 356347546), ErrReminderNotFound)
	})
}

func (s *storageTestSuite) Test_storage_SaveReminder() {
	s.Run("success", func() {
		reminder := domain.Reminder{
			ChatID:       1346,
			UserID:       7658,
			Text:         "Demand idaho agree reservoir may fisheries completion, baseline upon actions bond towards insurance trading, replacing spiritual.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}
		id, err := s.storage.SaveReminder(context.TODO(), reminder)
		s.NoError(err)

		actRemidner := s.mustGetReminder(id)
		requireEqualRemindersList(s.Require(), []domain.Reminder{reminder}, []domain.Reminder{actRemidner})
	})

	s.Run("success: createdAt and modifiedAt are not set", func() {
		reminder := domain.Reminder{
			ChatID:       1346,
			UserID:       7658,
			Text:         "Demand idaho agree reservoir may fisheries completion, baseline upon actions bond towards insurance trading, replacing spiritual.",
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}
		id, err := s.storage.SaveReminder(context.TODO(), reminder)
		s.NoError(err)

		actRemidner := s.mustGetReminder(id)
		s.NotZero(actRemidner.CreatedAt)
		s.NotZero(actRemidner.ModifiedAt)
	})
}

func (s *storageTestSuite) Test_storage_SetReminderStatus() {
	s.Run("success", func() {
		reminder := domain.Reminder{
			ChatID:       1346,
			UserID:       7658,
			Text:         "Demand idaho agree reservoir may fisheries completion, baseline upon actions bond towards insurance trading, replacing spiritual.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}

		id, err := s.storage.SaveReminder(context.TODO(), reminder)
		s.NoError(err)

		s.NoError(s.storage.SetReminderStatus(context.TODO(), id, domain.ReminderStatusDone))

		actReminder := s.mustGetReminder(id)
		s.Equal(domain.ReminderStatusDone, actReminder.Status)
		s.Greater(actReminder.ModifiedAt, reminder.RemindAt)
	})

	s.Run("error: not found", func() {
		s.ErrorIs(s.storage.SetReminderStatus(context.TODO(), 123124, domain.ReminderStatusDone), ErrReminderNotFound)
	})
}

func (s *storageTestSuite) Test_storage_UpdateReminder() {
	s.Run("success", func() {
		reminder := domain.Reminder{
			ChatID:       1346,
			UserID:       7658,
			Text:         "Demand idaho agree reservoir may fisheries completion, baseline upon actions bond towards insurance trading, replacing spiritual.",
			CreatedAt:    timeNowUTC().Truncate(1 * time.Minute),
			ModifiedAt:   timeNowUTC().Truncate(1 * time.Minute),
			RemindAt:     timeNowUTC().Truncate(1 * time.Minute),
			Status:       domain.ReminderStatusPending,
			AttemptsLeft: 3,
		}

		id, err := s.storage.SaveReminder(context.TODO(), reminder)
		s.NoError(err)

		reminder.Status = domain.ReminderStatusDone
		reminder.AttemptsLeft = 10
		reminder.RemindAt = timeNowUTC().Truncate(1 * time.Minute)
		reminder.ModifiedAt = timeNowUTC().Truncate(1 * time.Minute)
		reminder.ID = id

		s.NoError(s.storage.UpdateReminder(context.TODO(), reminder))

		actReminder := s.mustGetReminder(id)
		s.Equal(reminder, actReminder)
	})

	s.Run("error: not found", func() {
		s.ErrorIs(s.storage.UpdateReminder(context.TODO(), domain.Reminder{ID: 35689}), ErrReminderNotFound)
	})
}

func requireEqualRemindersList(r *require.Assertions, exp, act []domain.Reminder) {
	slices.SortFunc(exp, func(a, b domain.Reminder) int {
		return a.ModifiedAt.Compare(b.ModifiedAt)
	})
	slices.SortFunc(act, func(a, b domain.Reminder) int {
		return a.ModifiedAt.Compare(b.ModifiedAt)
	})

	r.Equal(len(exp), len(act))

	for i, expEl := range exp {
		actEl := act[i]
		expEl.ID = actEl.ID // do not compare generated IDs
		r.Equal(expEl, actEl)
	}
}

func (s *storageTestSuite) mustGetReminder(id int64) domain.Reminder {
	reminder, err := s.getGetReminder(id)
	if err != nil {
		s.FailNow(err.Error())
	}
	return reminder
}

func (s *storageTestSuite) getGetReminder(id int64) (domain.Reminder, error) {
	var reminder domain.Reminder
	if err := s.storage.db.Get(&reminder, `SELECT * FROM reminders;`, id); err != nil {
		return domain.Reminder{}, err
	}
	return reminder, nil
}

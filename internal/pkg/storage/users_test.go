package storage

import (
	"context"
	"time"

	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

func (s *storageTestSuite) Test_storage_SaveUser() {
	s.Run("success: user does not exist", func() {
		// ARRANGE
		user := domain.User{
			ID:         3,
			Name:       "Angelique Henke",
			Status:     domain.UserStatusActive,
			CreatedAt:  timeNowUTC(),
			ModifiedAt: timeNowUTC().Add(1 * time.Hour),
		}

		// ACT
		s.NoError(s.storage.SaveUser(context.TODO(), user))

		// ASSERT
		actUser := s.mustGetUser(user.ID)
		s.Equal(user, actUser)
	})

	s.Run("success: modifiedAt and createdAt are not set", func() {
		// ARRANGE
		user := domain.User{
			ID:     3,
			Name:   "Angelique Henke",
			Status: domain.UserStatusActive,
		}

		// ACT
		s.NoError(s.storage.SaveUser(context.TODO(), user))

		// ASSERT
		actUser := s.mustGetUser(user.ID)
		s.NotZero(actUser.CreatedAt)
		s.NotZero(actUser.ModifiedAt)
	})

	s.Run("error: user already exist", func() {
		// ARRANGE
		user := domain.User{
			ID:         4,
			Name:       "Angelique Henke",
			Status:     domain.UserStatusActive,
			CreatedAt:  timeNowUTC(),
			ModifiedAt: timeNowUTC().Add(1 * time.Hour),
		}

		// ACT
		s.NoError(s.storage.SaveUser(context.TODO(), user))

		// ASSERT
		s.ErrorIs(s.storage.SaveUser(context.TODO(), user), ErrUserAlreadyExists)
	})
}

func (s *storageTestSuite) Test_storage_SetUserStatus() {
	s.Run("success: user exists", func() {
		// ARRANGE
		user := domain.User{
			ID:         5748,
			Name:       "Angelique Henke",
			Status:     domain.UserStatusActive,
			CreatedAt:  timeNowUTC(),
			ModifiedAt: timeNowUTC(),
		}
		s.NoError(s.storage.SaveUser(context.TODO(), user))

		// ACT
		s.NoError(s.storage.SetUserStatus(context.TODO(), user.ID, domain.UserStatusInactive))

		// ASSERT
		actUser := s.mustGetUser(user.ID)
		s.Equal(domain.UserStatusInactive, actUser.Status)
		s.Greater(actUser.ModifiedAt, user.ModifiedAt)
	})

	s.Run("success: user does not exist", func() {
		s.ErrorIs(s.storage.SetUserStatus(context.TODO(), 10, domain.UserStatusInactive), ErrUserNotFound)
	})
}

func (s *storageTestSuite) mustGetUser(id int64) domain.User {
	var user domain.User
	if err := s.storage.db.Get(&user, `SELECT * FROM users;`, id); err != nil {
		s.FailNow(err.Error())
	}
	return user
}

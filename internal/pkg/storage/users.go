package storage

import (
	"context"
	"errors"
	"fmt"

	log "github.com/go-pkgz/lgr"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

var (
	// ErrUserAlreadyExists - user is already exist.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrUserNotFound - user is not found.
	ErrUserNotFound = errors.New("user is not found")
)

// SaveUser - saves user.
func (s *Storage) SaveUser(ctx context.Context, user domain.User) error {
	now := timeNowUTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.ModifiedAt.IsZero() {
		user.ModifiedAt = now
	}

	const query = `INSERT INTO users(
            id
            , name
            , status      
            , created_at      
            , modified_at      
	) VALUES ($1, $2, $3, $4, $5)`

	if _, err := s.db.ExecContext(ctx, query, user.ID, user.Name, user.Status, user.CreatedAt, user.ModifiedAt); err != nil {
		switch {
		case isAlreadyExistsError(err):
			return fmt.Errorf("failed to save user %s: %w", user, ErrUserAlreadyExists)
		default:
			return fmt.Errorf("failed to save user %s: %w", user, err)
		}
	}

	log.Printf("[INFO] saved new user %d", user.ID)

	return nil
}

// SetUserStatus - set's user status by user id.
func (s *Storage) SetUserStatus(ctx context.Context, id int64, status domain.UserStatus) error {
	const query = `UPDATE users SET status = $1, modified_at = $2 WHERE id = $3;`

	res, err := s.db.ExecContext(ctx, query, status, timeNowUTC(), id)
	if err != nil {
		return fmt.Errorf("failed to set user status to %s: %w", status, err)
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return fmt.Errorf("failed to set user status to %s: %w", status, ErrUserNotFound)
	}

	log.Printf("[INFO] set user %d status to %s", id, status)

	return nil
}

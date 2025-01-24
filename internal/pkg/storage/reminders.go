package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

// ErrReminderNotFound - reminder is not found
var ErrReminderNotFound = errors.New("reminder is not found")

// GetMyReminders - returns reminders by user id and chat id.
func (s *Storage) GetMyReminders(ctx context.Context, userID, chatID int64) ([]domain.Reminder, error) {
	const query = `
		SELECT
		    id
			, chat_id
			, user_id
			, text
			, created_at
			, modified_at
			, remind_at
			, status
			, attempts_left
		FROM reminders	
		WHERE user_id = $1
			AND chat_id = $2
			AND status = 'pending'
		ORDER BY remind_at`

	var reminders []domain.Reminder
	if err := s.db.SelectContext(ctx, &reminders, query, userID, chatID); err != nil {
		return nil, fmt.Errorf("failed to get my reminders: %w", err)
	}

	log.Printf("[DEBUG] got %d reminders for user %d", len(reminders), userID)

	return reminders, nil
}

// RemoveReminder - removes reminder by id.
func (s *Storage) RemoveReminder(ctx context.Context, id int64) error {
	const query = `DELETE FROM reminders WHERE id = $1;`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to remove reminder %d: %w", id, err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("failed to remove reminder %d: %w", id, ErrReminderNotFound)
	}

	log.Printf("[INFO] removed reminder %d", id)

	return nil
}

// SaveReminder - saves reminder.
func (s *Storage) SaveReminder(ctx context.Context, reminder domain.Reminder) (int64, error) {
	now := timeNowUTC()
	if reminder.CreatedAt.IsZero() {
		reminder.CreatedAt = now
	}
	if reminder.ModifiedAt.IsZero() {
		reminder.ModifiedAt = now
	}

	const query = `
		INSERT INTO reminders(
			chat_id
			, user_id
			, text
			, created_at
			, modified_at
			, remind_at
			, status
			, attempts_left             
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;`

	if err := s.db.GetContext(ctx, &reminder.ID, query,
		reminder.ChatID,
		reminder.UserID,
		reminder.Text,
		reminder.CreatedAt,
		reminder.ModifiedAt,
		reminder.RemindAt,
		reminder.Status,
		reminder.AttemptsLeft,
	); err != nil {
		return 0, fmt.Errorf("failed to save reminder %s: %w", reminder, err)
	}

	log.Printf("[INFO] saved reminder %s", reminder)

	return reminder.ID, nil
}

// UpdateReminder - updates reminder.
func (s *Storage) UpdateReminder(ctx context.Context, reminder domain.Reminder) error {
	if reminder.ModifiedAt.IsZero() {
		reminder.ModifiedAt = timeNowUTC()
	}

	const query = `
		UPDATE reminders
		SET status = $1
		    , attempts_left = $2
		    , remind_at = $3
			, modified_at = $4
		WHERE id = $5;`

	res, err := s.db.ExecContext(ctx, query,
		reminder.Status,
		reminder.AttemptsLeft,
		reminder.RemindAt,
		reminder.ModifiedAt,
		reminder.ID,
	)
	if err != nil {
		return err
	}

	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrReminderNotFound
	}

	log.Printf("[INFO] updated reminder %s", reminder)

	return nil
}

// GetPendingReminders - returns reminders in [domain.ReminderStatusPending] status for active users.
func (s *Storage) GetPendingReminders(ctx context.Context, limit int64) ([]domain.Reminder, error) {
	const query = `
		SELECT
		    r.id
			, r.chat_id
			, r.user_id
			, r.text
			, r.created_at
			, r.modified_at
			, r.remind_at
			, r.status
			, r.attempts_left
		FROM reminders r
		JOIN users u ON r.user_id = u.id
		WHERE r.status = 'pending'
			AND r.remind_at < $1
			AND r.attempts_left > 0
			AND u.status = 'active';`

	var reminders []domain.Reminder

	if err := s.db.SelectContext(ctx, &reminders, query, timeNowUTC(), limit); err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] got %d pending reminders", len(reminders))

	return reminders, nil
}

// SetReminderStatus - set's reminder status by id.
func (s *Storage) SetReminderStatus(ctx context.Context, id int64, status domain.ReminderStatus) error {
	const query = `UPDATE reminders SET status = $1, modified_at = $2 WHERE id = $3;`

	res, err := s.db.ExecContext(ctx, query, status, timeNowUTC(), id)
	if err != nil {
		return fmt.Errorf("failed to reminder %d status to %s: %w", id, status, err)
	}

	if affected, _ := res.RowsAffected(); affected == 0 {
		return fmt.Errorf("failed to reminder %d status to %s: %w", id, status, ErrReminderNotFound)
	}

	log.Printf("[INFO] set reminder %d status to %s", id, status)

	return nil
}

// DelayReminder - delays reminder by id. Reminder will be fired at remindAt time.
func (s *Storage) DelayReminder(ctx context.Context, id int64, remindAt time.Time) error {
	const query = "UPDATE reminders SET remind_at = $1, attempts_left = $2, modified_at = $3 WHERE id = $4 AND status = 'pending';"

	res, err := s.db.ExecContext(ctx, query, remindAt, domain.DefaultAttemptsLeft, timeNowUTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delay reminder %d: %w", id, err)
	}

	if affected, _ := res.RowsAffected(); affected == 0 {
		return fmt.Errorf("failed to delay reminder %d: %w", id, ErrReminderNotFound)
	}

	log.Printf("[INFO] delayed reminder [ID: %d, RemindAt: %s, AttemptsLeft: %d]", id, remindAt, domain.DefaultAttemptsLeft)

	return nil
}

func (s *Storage) GetReminderByID(ctx context.Context, id int64) (domain.Reminder, error) {
	const query = `
		SELECT
		    r.id
			, r.chat_id
			, r.user_id
			, r.text
			, r.created_at
			, r.modified_at
			, r.remind_at
			, r.status
			, r.attempts_left
		FROM reminders r
		WHERE r.id = $1;`

	var reminder domain.Reminder
	if err := s.db.GetContext(ctx, &reminder, query, id); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.Reminder{}, fmt.Errorf("failed to get reminder by id %d: %w", id, ErrBotStateNotFound)
		default:
			return domain.Reminder{}, fmt.Errorf("failed to get reminder by id %d: %w", id, err)
		}
	}

	log.Printf("[DEBUG] got reminder by %s", reminder)

	return reminder, nil
}

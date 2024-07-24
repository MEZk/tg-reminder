package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	log "github.com/go-pkgz/lgr"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
)

var (
	// ErrBotStateNotFound - bot state is not found.
	ErrBotStateNotFound = errors.New("bot state is not found")
	// ErrBotStateAlreadyExists - bot state is already exists.
	ErrBotStateAlreadyExists = errors.New("bot state already exists")
)

// SaveBotState - saves bot state.
func (s *Storage) SaveBotState(ctx context.Context, state domain.BotState) error {
	if state.ModifiedAt.IsZero() {
		state.ModifiedAt = timeNowUTC()
	}

	const query = `INSERT INTO bot_states(
            user_id
            , name
            , context
            , modified_at      
	) VALUES ($1, $2, $3, $4)
	ON CONFLICT DO UPDATE SET
		name = $2
		, context = $3
		, modified_at = $4;`

	if _, err := s.db.ExecContext(ctx, query, state.UserID, state.Name, state.Context, state.ModifiedAt); err != nil {
		return fmt.Errorf("failed to save bot state %s: %w", state, err)
	}

	log.Printf("[INFO] saved bot state %s", state)

	return nil
}

// GetBotState - returns bot state by user id.
func (s *Storage) GetBotState(ctx context.Context, userID int64) (domain.BotState, error) {
	const query = `
		SELECT
		    user_id
			, name
			, context
			, modified_at
		FROM bot_states
		WHERE user_id = $1;`

	var state domain.BotState
	if err := s.db.GetContext(ctx, &state, query, userID); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.BotState{}, fmt.Errorf("failed to get bot state for user %d: %w", userID, ErrBotStateNotFound)
		default:
			return domain.BotState{}, fmt.Errorf("failed to get bot state for user %d: %w", userID, err)
		}
	}

	log.Printf("[DEBUG] got bot state %s", state)

	return state, nil
}

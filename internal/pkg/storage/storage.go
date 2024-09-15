package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"modernc.org/sqlite"
)

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
}

// Storage - storage.
type Storage struct {
	db *sqlx.DB
}

// NewSqllite creates a new sqlite Storage.
func NewSqllite(db *sqlx.DB, migrationsDir string) (*Storage, error) {
	if err := goose.SetDialect(string(goose.DialectSQLite3)); err != nil {
		return nil, fmt.Errorf("can't set sql dialect: %w", err)
	}
	if err := goose.Up(db.DB, migrationsDir); err != nil {
		return nil, fmt.Errorf("failed to up migrations: %w", err)
	}

	return &Storage{db: db}, nil
}

func isAlreadyExistsError(err error) bool {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		const codeAlreadyExists = 1555
		return sqliteErr.Code() == codeAlreadyExists
	}

	return false
}

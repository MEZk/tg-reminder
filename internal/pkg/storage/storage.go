package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"modernc.org/sqlite"
	_ "modernc.org/sqlite" // sqlite driver loaded here
)

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
}

type storage struct {
	db *sqlx.DB
}

// NewSqllite creates a new sqlite storage.
func NewSqllite(file, migrationsDir string) (*storage, error) {
	db, err := sqlx.Connect("sqlite", file)
	if err != nil {
		return nil, err
	}

	goose.SetDialect(string(goose.DialectSQLite3))
	if err = goose.Up(db.DB, migrationsDir); err != nil {
		return nil, fmt.Errorf("failed to up migrations: %w", err)
	}

	return &storage{db: db}, nil
}

func isAlreadyExistsError(err error) bool {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		const codeAlreadyExists = 1555
		return sqliteErr.Code() == codeAlreadyExists
	}

	return false
}

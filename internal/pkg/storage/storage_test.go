package storage

import (
	"path"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

const testDB = "bot_test.db"

type storageTestSuite struct {
	suite.Suite
	dbPath  string
	storage *Storage
}

//nolint:paralleltest // we clean db after each test, see TearDownSubTest
func Test_StorageTestSuite(t *testing.T) {
	suite.Run(t, new(storageTestSuite))
}

func (s *storageTestSuite) SetupSuite() {
	timeNowUTC = func() time.Time {
		return time.Now().UTC()
	}

	s.dbPath = path.Join(s.T().TempDir(), testDB)

	db, err := sqlx.Connect("sqlite", s.dbPath)
	if err != nil {
		s.FailNow(err.Error())
	}

	store, err := NewSqllite(db, "../../../migrations")
	if err != nil {
		s.FailNow(err.Error())
	}

	s.storage = store

	s.T().Logf("[INFO] test database %s created in %s", testDB, s.dbPath)
}

func (s *storageTestSuite) TearDownSuite() {
	s.storage.db.Close()
}

func (s *storageTestSuite) TearDownSubTest() {
	if _, err := s.storage.db.Exec(`
		DELETE FROM reminders;
		DELETE FROM users;
		DELETE FROM bot_states;
	`); err != nil {
		s.FailNow(err.Error())
	}
}

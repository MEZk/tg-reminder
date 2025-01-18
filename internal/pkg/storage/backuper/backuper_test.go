package backuper

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestBackuper_deleteOldBackups(t *testing.T) {
	t.Parallel()

	t.Run("success: remove old backup file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		backupFilename := tmpDir + "/file_must_be_removed1" + backupFileExtension

		r := require.New(t)

		if _, err := os.Create(backupFilename); err != nil {
			r.Fail(fmt.Sprintf("can't create backup file %s: %v", backupFilename, err))
		}

		b, err := New(nil, tmpDir, 0, 0)
		r.NoError(err)

		r.NoError(b.deleteOldBackups())

		_, err = os.Stat(backupFilename)
		r.True(os.IsNotExist(err))
	})

	t.Run("success: backup dir contains backup and txt file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		backupFilename := tmpDir + "/file_must_be_removed1" + backupFileExtension

		r := require.New(t)

		if _, err := os.Create(backupFilename); err != nil {
			r.Fail(fmt.Sprintf("can't create backup file %s: %v", backupFilename, err))
		}

		txtFilename := tmpDir + "/txtfile.txt"
		if _, err := os.Create(txtFilename); err != nil {
			r.Fail(fmt.Sprintf("can't create txt file %s: %v", backupFilename, err))
		}

		b, err := New(nil, tmpDir, 0, 0)
		r.NoError(err)

		r.NoError(b.deleteOldBackups())

		_, err = os.Stat(backupFilename)
		r.True(os.IsNotExist(err))

		_, err = os.Stat(txtFilename)
		r.NoError(err)
	})

	t.Run("success: do not remove old backup file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		backupFilename := tmpDir + "/file_must_be_removed2" + backupFileExtension

		r := require.New(t)

		if _, err := os.Create(backupFilename); err != nil {
			r.Fail(fmt.Sprintf("can't create backup file %s: %v", backupFilename, err))
		}

		b, err := New(nil, tmpDir, 0, 1*time.Hour)
		r.NoError(err)

		r.NoError(b.deleteOldBackups())

		stat, err := os.Stat(backupFilename)
		r.NoError(err)
		r.Equal("file_must_be_removed2.backup", stat.Name())
	})

	t.Run("error: backup dir dose not exist", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		r := require.New(t)

		b, err := New(nil, tmpDir, 0, 0)
		r.NoError(err)

		r.NoError(os.RemoveAll(tmpDir))

		err = b.deleteOldBackups()
		r.Error(err)
		r.True(os.IsNotExist(err))
	})
}

func TestBackuper_Run(t *testing.T) {
	t.Parallel()

	t.Run("success: create db backup", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		backupDir := path.Join(tmpDir, "backup")
		dbFile := tmpDir + "/test_backup.db"

		r := require.New(t)

		db, err := sqlx.Connect("sqlite", dbFile)
		if err != nil {
			r.FailNow(err.Error())
		}
		t.Cleanup(func() {
			db.Close()
		})

		b, err := New(db, backupDir, 100*time.Millisecond, 0)
		r.NoError(err)

		timeNow := timeNowUTC().Truncate(time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		b.Run(ctx)

		backupDirEntries, err := os.ReadDir(backupDir)
		r.NoError(err)
		r.Len(backupDirEntries, 1)

		backupTimePrefix, ok := strings.CutSuffix(backupDirEntries[0].Name(), backupFileExtension)
		r.True(ok)

		backupTime, err := time.Parse(time.RFC3339, backupTimePrefix)
		r.NoError(err)
		r.Equal(timeNow, backupTime.Truncate(time.Minute))
	})

	t.Run("success: create db backup, failed to remove old backups as backup directory does not exist", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		backupDir := path.Join(tmpDir, "backup1234")
		dbFile := tmpDir + "/test_backup.db"

		r := require.New(t)

		db, err := sqlx.Connect("sqlite", dbFile)
		if err != nil {
			r.FailNow(err.Error())
		}
		t.Cleanup(func() {
			db.Close()
		})

		b, err := New(db, backupDir, 100*time.Millisecond, 0)
		r.NoError(err)

		r.NoError(os.RemoveAll(tmpDir))

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		b.Run(ctx)

		_, err = os.ReadDir(backupDir)
		r.ErrorIs(err, fs.ErrNotExist)
	})
}

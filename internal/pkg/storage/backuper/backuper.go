package backuper

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/jmoiron/sqlx"
)

// Backuper makes db backups.
type Backuper struct {
	db        *sqlx.DB
	backupDir string
	retention time.Duration
	interval  time.Duration
}

// New creates [Backuper].
func New(db *sqlx.DB, backupDir string, interval, retention time.Duration) (*Backuper, error) {
	const fMode = fs.FileMode(0o750) // User:rwx Group:r-x World:--- (i.e. World: no access)
	if err := os.MkdirAll(backupDir, fMode); err != nil {
		return nil, fmt.Errorf("failed to create backup directory %s: %w", backupDir, err)
	}

	return &Backuper{
		db:        db,
		backupDir: backupDir,
		retention: retention,
		interval:  interval,
	}, nil
}

const backupFileExtension = ".backup"

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
}

// Run starts [Backuper].
// Backuper deletes old backup files in accordance with retention interval and creates new backups in accordance with backup interval.
func (b *Backuper) Run(ctx context.Context) {
	log.Printf("[INFO] backuper started, backup interval %s, retention %s, backupDir %s", b.interval, b.retention, b.backupDir)

	ticker := time.NewTicker(b.interval)

	for {
		select {

		case <-ctx.Done():
			log.Printf("[INFO] backuper is shutting down")
			return

		case <-ticker.C:
			log.Printf("[DEBUG] backuper starts doing backup")

			if err := b.deleteOldBackups(); err != nil {
				log.Printf("[ERROR] failed to delete old backups: %v", err)
			}

			backupFilename := fmt.Sprintf("%s/%s%s", b.backupDir, timeNowUTC().Format(time.RFC3339), backupFileExtension)

			if _, err := b.db.ExecContext(ctx, "VACUUM INTO $1", backupFilename); err != nil {
				log.Printf("[ERROR] failed to do backup %s: %v", backupFilename, err)
				continue
			}

			log.Printf("[INFO] backuper finished doing backup %s", backupFilename)
		}
	}
}

func (b *Backuper) deleteOldBackups() error {
	dirEntries, err := os.ReadDir(b.backupDir)
	if err != nil {
		return err
	}

	deleteDate := timeNowUTC().Add(-b.retention)

	for _, entry := range dirEntries {
		if entry.IsDir() || path.Ext(entry.Name()) != backupFileExtension {
			continue
		}

		var info fs.FileInfo
		if info, err = entry.Info(); err != nil {
			return err
		}

		if info.ModTime().Before(deleteDate) {
			if err = os.Remove(b.backupDir + "/" + info.Name()); err != nil {
				return fmt.Errorf("failed to remove backup %s: %v", info.Name(), err)
			}

			log.Printf("[INFO] backuper removed old backup %s", info.Name())
		}
	}

	return nil
}

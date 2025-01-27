package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fatih/color"
	log "github.com/go-pkgz/lgr"
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	v2 "github.com/mezk/tg-reminder/internal/pkg/bot/v2"
	"github.com/mezk/tg-reminder/internal/pkg/storage"
	"github.com/mezk/tg-reminder/internal/pkg/storage/backuper"
)

const (
	envDBFile                 = "DB_FILE"                   // database file path
	envMigrations             = "MIGRATIONS"                // migration directory for goose
	envDebug                  = "DEBUG"                     // whether to print debug logs
	envTelegramAPIToken       = "TELEGRAM_APITOKEN"         // Telegram API token, received from Botfather
	envTelegramBotAPIEndpoint = "TELEGRAM_BOT_API_ENDPOINT" // Telegram API Bot endpoint
	envBackupRetention        = "BACKUP_RETENTION"          // backup retention interval
	envBackupInterval         = "BACKUP_INTERVAL"           // backup interval
	envBackupDir              = "BACKUP_DIR"                // backup files directory
)

var revision = "local"

func main() {
	fmt.Printf("tg-reminder %s\n", revision)

	if err := execute(); err != nil {
		log.Printf("[ERROR] %v", err)
		os.Exit(1)
	}
}

func execute() error {
	dbFile := os.Getenv(envDBFile)

	migrationsDir := "/srv/db/migrations" // default location
	if dir := os.Getenv(envMigrations); dir != "" {
		migrationsDir = os.Getenv(envMigrations)
	}

	debug, err := strconv.ParseBool(os.Getenv(envDebug))
	if err != nil {
		return fmt.Errorf("fail to parse %s env variable: %w", envDebug, err)
	}

	tgAPIToken := os.Getenv(envTelegramAPIToken)
	masked := []string{tgAPIToken}
	if err = setupLog(debug, masked...); err != nil {
		return fmt.Errorf("fail to setup logger: %w", err)
	}

	log.Printf("[INFO start bot [Revision: %s, DBFile: %s, MigrationsDir: %s, Debug: %t]", revision, dbFile, migrationsDir, debug)

	botAPIEndpoint := os.Getenv(envTelegramBotAPIEndpoint)
	if botAPIEndpoint == "" {
		botAPIEndpoint = tbapi.APIEndpoint
	}

	botAPI, err := tbapi.NewBotAPIWithAPIEndpoint(tgAPIToken, botAPIEndpoint)
	if err != nil {
		return fmt.Errorf("can't connect to telegram bot api: %w", err)
	}
	botAPI.Debug = debug

	db, err := sqlx.Connect("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("can't connect to database: %w", err)
	}

	store, err := storage.NewSqllite(db, migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to connect to sqlite %s: %v", dbFile, err)
	}

	bot, err := v2.New(store, tgAPIToken)
	if err != nil {
		return fmt.Errorf("can't create bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
		bot.Stop()
	}()

	if backupDir := os.Getenv(envBackupDir); backupDir != "" {
		var backupRetentionInterval time.Duration
		if backupRetentionInterval, err = time.ParseDuration(os.Getenv(envBackupRetention)); err != nil {
			return fmt.Errorf("can't parse backup retention interval: %w", err)
		}

		var backupInterval time.Duration
		if backupInterval, err = time.ParseDuration(os.Getenv(envBackupInterval)); err != nil {
			return fmt.Errorf("can't parse backup interval: %w", err)
		}

		var backup *backuper.Backuper
		if backup, err = backuper.New(db, backupDir, backupInterval, backupRetentionInterval); err != nil {
			return fmt.Errorf("can't create backuper: %w", err)
		}
		// backuper starts in background goroutine
		go func() {
			backup.Run(ctx)
		}()
	}

	// notificationSender := notifier.New(tgMessageSender, store, 1*time.Minute)
	// notifications sender starts in background goroutine
	// go func() {
	// 	notificationSender.Run(ctx)
	// }()

	// Start is a blocking call
	bot.Start()

	return nil
}

func setupLog(dbg bool, secrets ...string) error {
	logOpts := []log.Option{log.Msec, log.LevelBraces, log.StackTraceOnError}
	if dbg {
		logOpts = []log.Option{log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces, log.StackTraceOnError}
	}

	colorizer := log.Mapper{
		ErrorFunc:  func(s string) string { return color.New(color.FgHiRed).Sprint(s) },
		WarnFunc:   func(s string) string { return color.New(color.FgRed).Sprint(s) },
		InfoFunc:   func(s string) string { return color.New(color.FgYellow).Sprint(s) },
		DebugFunc:  func(s string) string { return color.New(color.FgWhite).Sprint(s) },
		CallerFunc: func(s string) string { return color.New(color.FgBlue).Sprint(s) },
		TimeFunc:   func(s string) string { return color.New(color.FgCyan).Sprint(s) },
	}
	logOpts = append(logOpts, log.Map(colorizer))

	if len(secrets) > 0 {
		logOpts = append(logOpts, log.Secret(secrets...))
	}

	log.SetupStdLogger(logOpts...)
	log.Setup(logOpts...)

	if err := tbapi.SetLogger(log.ToStdLogger(log.Default(), "DEBUG tbapi ----")); err != nil {
		return err
	}

	return nil
}

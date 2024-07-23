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
	"github.com/mezk/tg-reminder/internal/pkg/bot"
	"github.com/mezk/tg-reminder/internal/pkg/listener"
	"github.com/mezk/tg-reminder/internal/pkg/notifier"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
	"github.com/mezk/tg-reminder/internal/pkg/storage"
)

const (
	envDbFile                 = "DB_FILE"                   // database file path
	envMigrations             = "MIGRATIONS"                // migration folders for goose
	envDebug                  = "DEBUG"                     // whether to print debug logs
	envTelegramAPIToken       = "TELEGRAM_APITOKEN"         // Telegram API token, received from Botfather
	envTelegramBotAPIEndpoint = "TELEGRAM_BOT_API_ENDPOINT" // Telegram API Bot endpoint
)

var revision = "local"

func main() {
	fmt.Printf("tg-reminder %s\n", revision)

	dbFile := os.Getenv(envDbFile)
	migrationsFolder := os.Getenv(envMigrations)

	debug, err := strconv.ParseBool(os.Getenv(envDebug))
	if err != nil {
		panic(err)
	}

	tgAPIToken := os.Getenv(envTelegramAPIToken)
	masked := []string{tgAPIToken}
	setupLog(debug, masked...)

	log.Printf("[INFO start bot [Revision: %s, DBFile: %s, MigrationsFolder: %s, Debug: %t]", revision, dbFile, migrationsFolder, debug)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	if err = execute(ctx, dbFile, migrationsFolder, tgAPIToken, debug); err != nil {
		log.Printf("[ERROR] %v", err)
		os.Exit(1)
	}
}

func execute(ctx context.Context, dbFile, migrationsFolder, tgAPIToken string, debug bool) error {
	botAPIEndpoint := os.Getenv(envTelegramBotAPIEndpoint)
	if botAPIEndpoint == "" {
		botAPIEndpoint = tbapi.APIEndpoint
	}

	botAPI, err := tbapi.NewBotAPIWithAPIEndpoint(tgAPIToken, botAPIEndpoint)
	if err != nil {
		return fmt.Errorf("can't connect to telegram bot api: %w", err)
	}
	botAPI.Debug = debug

	store, err := storage.NewSqllite(dbFile, migrationsFolder)
	if err != nil {
		return fmt.Errorf("failed to connect to sqlite %s: %v", dbFile, err)
	}

	tgMessageSender := sender.New(botAPI)

	reminderBot := bot.New(tgMessageSender, store)

	tgUpdatesListener := listener.New(botAPI, reminderBot)

	notificationSender := notifier.New(tgMessageSender, store, 1*time.Minute)
	// notifications sender starts in background goroutine
	go func() {
		notificationSender.Run(ctx)
	}()

	// Listen is a blocking call
	tgUpdatesListener.Listen(ctx)
	return nil
}

func setupLog(dbg bool, secrets ...string) {
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
	tbapi.SetLogger(log.ToStdLogger(log.Default(), "DEBUG tbapi ----"))
}

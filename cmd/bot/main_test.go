package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testChatID   = 32467
	testUserID   = 345757
	testUserName = "JohnDoe"
)

func Test_execute(t *testing.T) {
	const (
		testAPIToken  = "e30637b8-8d51-457d-b8e4-0eee79e3cf5c"
		migrationsDir = "../../migrations"
		envForkedTest = "FORK"
		botID         = "/bot" + testAPIToken
	)

	t.Run("success: start bot, create db, make db backup, send getMe, getUpdates, sendMessage requests to Telegram", func(t *testing.T) {
		r := require.New(t)

		// ARRANGE
		tempDir := t.TempDir()
		f, err := os.CreateTemp(tempDir, "test-db")
		if err != nil {
			r.NoError(err)
		}

		var (
			dbFile      = f.Name()
			dbBackupDir = path.Join(tempDir, "db_backup")

			testUser = tbapi.User{ID: testUserID, UserName: testUserName}

			testUserJSON, _    = json.Marshal(testUser)
			testUpdatesJSON, _ = json.Marshal([]tbapi.Update{
				{
					Message: &tbapi.Message{
						Chat: &tbapi.Chat{ID: testChatID},
						Text: "/start",
						From: &testUser,
					},
				},
			})
			testMessageJSON, _ = json.Marshal(tbapi.Message{
				MessageID: 1,
				From:      &testUser,
			})

			hasUpdates  atomic.Bool
			hasMessages atomic.Bool
		)
		hasUpdates.Store(true)
		hasMessages.Store(true)

		tgAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			a := assert.New(t)

			switch {
			case strings.HasSuffix(req.URL.Path, "getMe"):
				a.Equal(botID+"/getMe", req.URL.Path)
				writeTgServerResp(t, w, tbapi.APIResponse{Ok: true, Result: testUserJSON})
			case strings.HasSuffix(req.URL.Path, "getUpdates"):
				a.Equal(botID+"/getUpdates", req.URL.Path)

				if hasUpdates.Load() {
					writeTgServerResp(t, w, tbapi.APIResponse{Ok: true, Result: testUpdatesJSON})
					hasUpdates.Store(false)
				} else {
					writeTgServerResp(t, w, tbapi.APIResponse{Ok: true, Result: nil})
				}
			case strings.HasSuffix(req.URL.Path, "sendMessage"):
				a.Equal(botID+"/sendMessage", req.URL.Path)

				var body []byte
				body, err = io.ReadAll(req.Body)
				a.NoError(err)

				var actMsg string
				actMsg, err = url.QueryUnescape(string(body))
				a.NoError(err)

				const expMsg = `chat_id=32467&disable_web_page_preview=true&entities=null&parse_mode=Markdown&text=*Привет,* @JohnDoe 👋

Теперь вы можете со мной работать.
Для справки 💁 используйте команду /help`

				a.Equal(expMsg, actMsg)

				if hasMessages.Load() {
					writeTgServerResp(t, w, tbapi.APIResponse{Ok: true, Result: testMessageJSON})
					hasMessages.Store(false)
				} else {
					writeTgServerResp(t, w, tbapi.APIResponse{Ok: true, Result: nil})
				}
			default:
				a.Failf("unknown request url path: %s", req.URL.Path)
			}
		}))

		t.Setenv(envTelegramBotAPIEndpoint, tgAPIServer.URL+"/bot%s/%s") // https://api.telegram.org/bot%s/%s
		t.Setenv(envMigrations, migrationsDir)
		t.Setenv(envDebug, "false")
		t.Setenv(envTelegramAPIToken, testAPIToken)
		t.Setenv(envDBFile, dbFile)
		t.Setenv(envBackupDir, dbBackupDir)
		t.Setenv(envBackupInterval, "700ms") // shpuld be less then sigIntTimeout, we expect exactly 1 backup before SIGINT signal
		t.Setenv(envBackupRetention, "2s")

		t.Cleanup(func() {
			f.Close()
			tgAPIServer.Close()
		})

		done := make(chan bool)
		// Send SIGINT after sigIntTimeout to stop execution
		const sigIntTimeout = 1 * time.Second
		go func() {
			ticker := time.NewTicker(sigIntTimeout)
			<-ticker.C
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			done <- true
			close(done)
		}()

		// ACT
		err = execute()

		// ASSERT
		r.ErrorIs(err, context.Canceled)
		r.True(<-done)
		checkDBStateAfterExecute(r, dbFile)

		backupDirEntries, err := os.ReadDir(dbBackupDir)
		r.NoError(err)
		r.Len(backupDirEntries, 1)
	})
}

func writeTgServerResp(t *testing.T, w http.ResponseWriter, resp tbapi.APIResponse) {
	t.Helper()

	w.WriteHeader(http.StatusOK)
	b, _ := json.Marshal(resp)
	fmt.Fprint(w, string(b))
}

func checkDBStateAfterExecute(r *require.Assertions, dbFile string) {
	db, err := sqlx.Connect("sqlite", dbFile)
	if err != nil {
		r.NoError(err)
	}

	const getTablesQuery = `
		SELECT name 
		FROM sqlite_schema
		WHERE 
			type ='table' AND 
			name NOT LIKE 'sqlite_%';`

	var tables []string
	r.NoError(db.Select(&tables, getTablesQuery))

	exTables := []string{
		"goose_db_version",
		"users",
		"reminders",
		"bot_states",
	}
	r.EqualValues(exTables, tables)

	const getUsersQuery = `SELECT * FROM users;`
	var user domain.User
	r.NoError(db.Get(&user, getUsersQuery))
	r.EqualValues(testUserID, user.ID)
	r.EqualValues(testUserName, user.Name)
	r.EqualValues(domain.UserStatusActive, user.Status)

	const getBotStateQuery = `SELECT * FROM bot_states WHERE user_id = $1;`
	var state domain.BotState
	r.NoError(db.Get(&state, getBotStateQuery, testUserID))
	r.EqualValues(domain.BotStateNameStart, state.Name)
}

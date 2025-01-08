package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/stretchr/testify/require"
)

const (
	testChatID    = 32467
	testUserID    = 345757
	testUserName  = "JohnDoe"
	testAPIToken  = "e30637b8-8d51-457d-b8e4-0eee79e3cf5c"
	migrationsDir = "../../migrations"
)

//nolint:testifylint // TODO: fix test, it does not work. See https://github.com/MEZk/tg-reminder/issues/25
func Test_main(t *testing.T) {
	t.Run("success: start bot, create db, send getMe, getUpdates, sendMessage requests", func(t *testing.T) {
		r := require.New(t)

		if os.Getenv("CTX_CANCEL") != "1" {
			cmd := exec.Command(os.Args[0], "-test.run=Test_main")
			cmd.Env = append(os.Environ(), "CTX_CANCEL=1")
			stdout, _ := cmd.StderrPipe()
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}

			gotBytes, _ := io.ReadAll(stdout)
			r.Contains(string(gotBytes), "[ERROR] context canceled")

			err := cmd.Wait()
			if e, ok := err.(*exec.ExitError); !ok || e.Success() {
				r.FailNowf("process ran with err %s, want exit status 1", err.Error())
			}
			return
		}

		f, err := os.CreateTemp("", "test-db")
		if err != nil {
			t.Fatal(err)
		}

		dbFilename := f.Name()

		defer func() {
			f.Close()
			os.Remove(dbFilename)
		}()

		t.Setenv(envMigrations, migrationsDir)
		t.Setenv(envDebug, "false")
		t.Setenv(envTelegramAPIToken, testAPIToken)
		t.Setenv(envDBFile, dbFilename)

		var (
			withUpdates        atomic.Bool
			testUser           = tbapi.User{ID: testUserID, UserName: testUserName}
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
		)
		withUpdates.Store(true)

		tgTestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			r.Equal(http.MethodPost, req.Method)
			switch {
			case strings.HasSuffix(req.URL.Path, "getMe"):
				r.Equal("/bot"+testAPIToken+"/getMe", req.URL.Path)
				writeTgServerResponse(w, tbapi.APIResponse{Ok: true, Result: testUserJSON})
			case strings.HasSuffix(req.URL.Path, "getUpdates"):
				r.Equal("/bot"+testAPIToken+"/getUpdates", req.URL.Path)
				if withUpdates.Load() {
					writeTgServerResponse(w, tbapi.APIResponse{Ok: true, Result: testUpdatesJSON})
				} else {
					writeTgServerResponse(w, tbapi.APIResponse{Ok: true, Result: nil})
				}
				// do not send update after first response
				withUpdates.Store(false)
			case strings.HasSuffix(req.URL.Path, "sendMessage"):
				r.Equal("/bot"+testAPIToken+"/sendMessage", req.URL.Path)
				checkSendMessageBody(r, req.Body)
				writeTgServerResponse(w, tbapi.APIResponse{Ok: true, Result: testMessageJSON})
			default:
				r.Failf("unknown request url path: %s", req.URL.Path)
			}
		}))
		defer tgTestServer.Close()

		// https://api.telegram.org/bot%s/%s
		apiEndpoint := tgTestServer.URL + "/bot%s/%s"
		t.Setenv(envTelegramBotAPIEndpoint, apiEndpoint)

		const sigIntTimeout = 1 * time.Second

		go func() {
			ticker := time.NewTicker(sigIntTimeout)
			<-ticker.C
			checkDBTables(r, dbFilename)
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()

		r.NotPanics(main)
	})
}

func writeTgServerResponse(w http.ResponseWriter, resp tbapi.APIResponse) {
	w.WriteHeader(http.StatusOK)
	b, _ := json.Marshal(resp)
	fmt.Fprint(w, string(b))
}

func checkDBTables(r *require.Assertions, dbFile string) {
	db, err := sqlx.Connect("sqlite", dbFile)
	if err != nil {
		r.NoError(err)
	}

	const query = `
		SELECT name 
		FROM sqlite_schema
		WHERE 
			type ='table' AND 
			name NOT LIKE 'sqlite_%';`

	var tables []string
	r.NoError(db.Select(&tables, query))

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

func checkSendMessageBody(r *require.Assertions, body io.ReadCloser) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(body)
	r.NoError(err)

	const expMsg = `chat_id=32467&disable_web_page_preview=true&entities=null&parse_mode=Markdown&text=Привет, @JohnDoe!
Теперь ты можешь со мной работать.
Для справки используй команду /help.`

	actMsg, err := url.QueryUnescape(buf.String())
	r.NoError(err)
	r.Equal(expMsg, actMsg)
}

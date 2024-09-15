package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	"github.com/stretchr/testify/assert"
)

const (
	testChatID       = 32467
	testUserID       = 345757
	testUserName     = "JohnDoe"
	testApiToken     = "e30637b8-8d51-457d-b8e4-0eee79e3cf5c"
	migrationsFolder = "../../migrations"
)

// TODO: fix test, it does not work. See https://github.com/MEZk/tg-reminder/issues/25
func Test_main(t *testing.T) {
	t.Run("success: start bot, create db, send getMe, getUpdates, sendMessage requests", func(t *testing.T) {
		a := assert.New(t)

		if os.Getenv("CTX_CANCEL") != "1" {
			cmd := exec.Command(os.Args[0], "-test.run=Test_main")
			cmd.Env = append(os.Environ(), "CTX_CANCEL=1")
			stdout, _ := cmd.StderrPipe()
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}

			gotBytes, _ := ioutil.ReadAll(stdout)
			a.Contains(string(gotBytes), "[ERROR] context canceled")

			err := cmd.Wait()
			if e, ok := err.(*exec.ExitError); !ok || e.Success() {
				a.FailNowf("process ran with err %s, want exit status 1", err.Error())
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

		t.Setenv(envMigrations, migrationsFolder)
		t.Setenv(envDebug, "false")
		t.Setenv(envTelegramAPIToken, testApiToken)
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
			a.Equal(http.MethodPost, req.Method)
			switch {
			case strings.HasSuffix(req.URL.Path, "getMe"):
				a.Equal("/bot"+testApiToken+"/getMe", req.URL.Path)
				writeTgServerResponse(w, tbapi.APIResponse{Ok: true, Result: testUserJSON}, http.StatusOK)
			case strings.HasSuffix(req.URL.Path, "getUpdates"):
				a.Equal("/bot"+testApiToken+"/getUpdates", req.URL.Path)
				if withUpdates.Load() {
					writeTgServerResponse(w, tbapi.APIResponse{Ok: true, Result: testUpdatesJSON}, http.StatusOK)
				} else {
					writeTgServerResponse(w, tbapi.APIResponse{Ok: true, Result: nil}, http.StatusOK)
				}
				// do not send update after first response
				withUpdates.Store(false)
			case strings.HasSuffix(req.URL.Path, "sendMessage"):
				a.Equal("/bot"+testApiToken+"/sendMessage", req.URL.Path)
				checkSendMessageBody(a, req.Body)
				writeTgServerResponse(w, tbapi.APIResponse{Ok: true, Result: testMessageJSON}, http.StatusOK)
			default:
				a.Failf("unknown request url path: %s", req.URL.Path)
			}
		}))
		defer tgTestServer.Close()

		// https://api.telegram.org/bot%s/%s
		apiEndpoint := tgTestServer.URL + "/bot%s/%s"
		t.Setenv(envTelegramBotAPIEndpoint, apiEndpoint)

		const sigIntTimeout = 1 * time.Second

		go func() {
			ticker := time.NewTicker(sigIntTimeout)
			select {
			case <-ticker.C:
				checkDBTables(a, dbFilename)
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}
		}()

		a.NotPanics(main)
	})
}

func writeTgServerResponse(w http.ResponseWriter, resp tbapi.APIResponse, status int) {
	w.WriteHeader(status)
	b, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(b))
}

func checkDBTables(a *assert.Assertions, dbFile string) {
	db, err := sqlx.Connect("sqlite", dbFile)
	if err != nil {
		a.NoError(err)
	}

	const query = `
		SELECT name 
		FROM sqlite_schema
		WHERE 
			type ='table' AND 
			name NOT LIKE 'sqlite_%';`

	var tables []string
	a.NoError(db.Select(&tables, query))

	exTables := []string{
		"goose_db_version",
		"users",
		"reminders",
		"bot_states",
	}
	a.EqualValues(exTables, tables)

	const getUsersQuery = `SELECT * FROM users;`
	var user domain.User
	a.NoError(db.Get(&user, getUsersQuery))
	a.EqualValues(testUserID, user.ID)
	a.EqualValues(testUserName, user.Name)
	a.EqualValues(domain.UserStatusActive, user.Status)

	const getBotStateQuery = `SELECT * FROM bot_states WHERE user_id = $1;`
	var state domain.BotState
	a.NoError(db.Get(&state, getBotStateQuery, testUserID))
	a.EqualValues(domain.BotStateNameStart, state.Name)
}

func checkSendMessageBody(a *assert.Assertions, body io.ReadCloser) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(body)
	a.NoError(err)

	const expMsg = `chat_id=32467&disable_web_page_preview=true&entities=null&parse_mode=Markdown&text=Привет, @JohnDoe!
Теперь ты можешь со мной работать.
Для справки используй команду /help.`

	actMsg, err := url.QueryUnescape(buf.String())
	a.NoError(err)
	a.Equal(expMsg, actMsg)
}

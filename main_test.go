package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type serverMock struct {
	s          *httptest.Server
	custom     map[string]any
	hooks      map[string]func(body []byte) any
	hooksCalls map[string]int
	updateIdx  int
	updates    []*models.Update
}

func (s *serverMock) Close() {
	s.s.Close()
}

func (s *serverMock) URL() string {
	return s.s.URL
}

type getUpdatesResponse struct {
	OK     bool             `json:"ok"`
	Result []*models.Update `json:"result"`
}

func (s *serverMock) handler(rw http.ResponseWriter, req *http.Request) {
	if req.URL.String() == "/bot/getMe" {
		_, err := rw.Write([]byte(`{"ok":true,"result":{}}`))
		if err != nil {
			panic(err)
		}
		return
	}
	if req.URL.String() == "/bot/editMessageText" {
		_, err := rw.Write([]byte(`{"ok":true,"result":{}}`))
		if err != nil {
			panic(err)
		}
		return
	}
	if req.URL.String() == "/bot/answerCallbackQuery" {
		_, err := rw.Write([]byte(`{"ok":true,"result":{}}`))
		if err != nil {
			panic(err)
		}
		return
	}
	if req.URL.String() == "/bot/getUpdates" {
		s.handlerGetUpdates(rw)
		return
	}

	reqBody, errReadBody := io.ReadAll(req.Body)
	if errReadBody != nil {
		panic(errReadBody)
	}
	defer req.Body.Close()

	hook, okHook := s.hooks[req.URL.String()]
	if okHook {
		s.hooksCalls[req.URL.String()]++
		resp, errData := json.Marshal(hook(reqBody))
		if errData != nil {
			panic(errData)
		}
		_, err := rw.Write(resp)
		if err != nil {
			panic(err)
		}
		return
	}

	d, ok := s.custom[req.URL.String()]
	if !ok {
		panic("answer not found for request: " + req.URL.String())
	}

	resp, errData := json.Marshal(d)
	if errData != nil {
		panic(errData)
	}
	_, err := rw.Write(resp)
	if err != nil {
		panic(err)
	}
}

func (s *serverMock) handlerGetUpdates(rw http.ResponseWriter) {
	if s.updateIdx >= len(s.updates) {
		_, err := rw.Write([]byte(`{"ok":true,"result":[]}`))
		if err != nil {
			panic(err)
		}
		return
	}

	s.updates[s.updateIdx].ID = int64(s.updateIdx + 1)

	r := getUpdatesResponse{
		OK:     true,
		Result: []*models.Update{s.updates[s.updateIdx]},
	}

	s.updateIdx++

	d, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	_, err = rw.Write(d)
	if err != nil {
		panic(err)
	}
}

func newServerMock() *serverMock {
	s := &serverMock{
		custom:     map[string]any{},
		hooks:      map[string]func([]byte) any{},
		hooksCalls: map[string]int{},
	}

	s.s = httptest.NewServer(http.HandlerFunc(s.handler))

	return s
}

func Test_shortenButtonText(t *testing.T) {
	type args struct {
		command  string
		name     string
		lastname string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				command:  "free-testing-1",
				name:     "",
				lastname: "",
			},
			want: "",
		},
		{
			name: "normal",
			args: args{
				command:  "free-testing-1",
				name:     "veryveryveryveryvery",
				lastname: "veryveryveryveryveryveryveryvery2",
			},
			want: "veryveryveryveryvery v.",
		},
		{
			name: "long",
			args: args{
				command:  "free-testing-1",
				name:     "veryveryveryveryvery1",
				lastname: "veryveryveryveryveryveryveryvery",
			},
			want: "veryveryveryveryvery1 v.",
		},
		{
			name: "long name",
			args: args{
				command:  "free-testing-1",
				name:     "veryveryveryveryveryveryveryveryveryveryveryvery",
				lastname: "veryveryveryveryveryveryveryvery",
			},
			want: "veryveryveryveryveryveryveryvery",
		},
		{
			name: "long lastname",
			args: args{
				command:  "free-testing-1",
				name:     "1veryveryveryveryveryveryveryver",
				lastname: "2veryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryvery",
			},
			want: "2veryveryveryveryveryveryveryver",
		},
		{
			name: "normal",
			args: args{
				command:  "free-testing-1",
				name:     "Firstname",
				lastname: "Firstname",
			},
			want: "Firstname Firstname",
		},
		{
			name: "normal cyrrilic",
			args: args{
				command:  "free-testing-1",
				name:     "ИмяОтчество",
				lastname: "Фамилия",
			},
			want: "UmяOтчecтвo Фamuлuя",
		},
		{
			name: "long cyrrilic name",
			args: args{
				command:  "free-testing-1",
				name:     "ИмяОтчество12345678901234567890123123123123123123123123123131231223131231231231312312123123123123123123123",
				lastname: "Фамилия",
			},
			want: "UmяOтчecтвo123456789012345678901",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortenUsername(tt.args.command, tt.args.name, tt.args.lastname)

			if got != tt.want {
				t.Errorf("shortenUsername(%s, %s, %s) = %v, want %v", tt.args.command, tt.args.name, tt.args.lastname, got, tt.want)
			}
		})
	}
}

func Test_handler(t *testing.T) {
	s := newServerMock()
	defer s.Close()

	b, err := bot.New("test_token", bot.WithServerURL(s.URL()), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("unexpected error %q", err)
	}

	type args struct {
		ctx    context.Context
		b      *bot.Bot
		update *models.Update
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "string",
			args: args{
				ctx: context.Background(),
				b:   b,
				update: &models.Update{
					CallbackQuery: &models.CallbackQuery{
						Data: "free-test",
						Message: &models.Message{
							Chat: models.Chat{
								ID: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "json",
			args: args{
				ctx: context.Background(),
				b:   b,
				update: &models.Update{
					CallbackQuery: &models.CallbackQuery{
						Data: `{"c": "free-test"}`,
						Message: &models.Message{
							Chat: models.Chat{
								ID: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "json notify",
			args: args{
				ctx: context.Background(),
				b:   b,
				update: &models.Update{
					CallbackQuery: &models.CallbackQuery{
						Data: `{"c": "⚡", "n": [1]}`,
						Message: &models.Message{
							Chat: models.Chat{
								ID: 1,
							},
						},
						Sender: models.User{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "bad json",
			args: args{
				ctx: context.Background(),
				b:   b,
				update: &models.Update{
					CallbackQuery: &models.CallbackQuery{
						Data: `{"c": "free-test"`,
						Message: &models.Message{
							Chat: models.Chat{
								ID: 1,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler(tt.args.ctx, tt.args.b, tt.args.update)
		})
	}
}

func Test_minifyJson(t *testing.T) {
	type args struct {
		input []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "good json",
			args: args{
				input: []byte(`{"c": "c"}`),
			},
			want: `{"c":"c"}`,
		},
		{
			name: "bad json",
			args: args{
				input: []byte(`{"c": "c"`),
			},
			want: `{"c": "c"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minifyJson(tt.args.input); got != tt.want {
				t.Errorf("minifyJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkStringLimit(t *testing.T) {
	type args struct {
		input string
		limit int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "63",
			args: args{
				input: "0123456789 0123456789 0123456789 0123456789 0123456789 01234567",
				limit: 64,
			},
			want: true,
		},
		{
			name: "65",
			args: args{
				input: "0123456789 0123456789 0123456789 0123456789 0123456789 0123456789",
				limit: 64,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkStringLimit(tt.args.input, tt.args.limit); got != tt.want {
				t.Errorf("checkStringLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}

//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

func mustBaccaratOutputJSON(msg string) string {
	out := &controller.BaccaratWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		BankerHand:    make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBaccaratOutputJSON: %v", err))
	}
	return string(b)
}

func baccaratIntPtr(v int) *int { return &v }

func TestBaccaratWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"bankerHand":[],"playerHandValue":0,"bankerHandValue":0,"phase":0,"chips":0,"betAmount":0,"betType":0,"result":0,"payout":0,"message":""}`
	expectedBody := mockOutput

	biMock := new(usecase.MockBaccaratInteractor)
	biMock.On("Reset").Return(mockOutput)
	biMock.On("Bet", 100, 0).Return(mockOutput)
	biMock.On("Bet", 100, 1).Return(mockOutput)
	biMock.On("Bet", 100, 2).Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BaccaratInteractorIF { return biMock }
	ctrl := controller.NewBaccaratWebController(factory)
	defer ctrl.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/baccarat/exec", ctrl.Exec),
	)
	api.SetApp(router)

	t.Run("quit q", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustBaccaratOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustBaccaratOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet player", func(t *testing.T) {
		input := controller.BaccaratWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Amount:       100,
			BetType:      baccaratIntPtr(0),
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet banker", func(t *testing.T) {
		input := controller.BaccaratWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			Amount:       100,
			BetType:      baccaratIntPtr(1),
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet tie", func(t *testing.T) {
		input := controller.BaccaratWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Amount:       100,
			BetType:      baccaratIntPtr(2),
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("action log", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("action log l", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBaccaratOutputJSON("Unsupported command."))
	})

	t.Run("param error empty", func(t *testing.T) {
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBaccaratOutputJSON("param error."))
	})

	t.Run("param error no command", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBaccaratOutputJSON("param error."))
	})

	t.Run("param error no session", func(t *testing.T) {
		var input controller.BaccaratWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":""}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/baccarat/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBaccaratOutputJSON("param error."))
	})

	t.Run("stop twice", func(t *testing.T) {
		ctrl2 := controller.NewBaccaratWebController(factory)
		ctrl2.Stop()
		ctrl2.Stop() // should not panic
	})
}

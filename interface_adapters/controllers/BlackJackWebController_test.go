package controllers_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers/usecases"
	uc "github.com/yuta-yoshinaga/go_trumpcards/usecases"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestBlackJackWebController_Method(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`
	bjiMock := new(usecases.MockBlackJackInteractor)
	bjiMock.On("Reset").Return(mockOutput).Times(2)
	bjiMock.On("Hit").Return(mockOutput)
	bjiMock.On("Stand").Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controllers.NewBlackJackWebController(factory)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/blackjack/exec", tbc.Exec),
	)
	api.SetApp(router)
	var jsonCase1 controllers.BlackJackWebInput
	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec h", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "h", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec hit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "hit", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec s", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "s", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec stand", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "stand", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"Unsupported command."}`)
	})
	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"param error."}`)
	})
	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"param error."}`)
	})
	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controllers.BlackJackWebInput{
			Command:   "reset",
			SessionId: strings.Repeat("a", controllers.SessionMaxIDLen+1),
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"param error."}`)
	})
	t.Run("failed Exec response empty", func(t *testing.T) {
		bjiMock.On("Reset").Return(``)
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"error."}`)
	})
}

func TestBlackJackWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"reset."}`
	mockA := new(usecases.MockBlackJackInteractor)
	mockA.On("Reset").Return(mockOutput)
	mockB := new(usecases.MockBlackJackInteractor)
	mockB.On("Reset").Return(mockOutput)

	callCount := 0
	isoController := controllers.NewBlackJackWebController(func() uc.BlackJackInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/blackjack/exec", isoController.Exec),
	)
	api.SetApp(router)

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controllers.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset")
		mockB.AssertNotCalled(t, "Reset")
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controllers.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset")
	})

	t.Run("session-A second call reuses mockA not mockB", func(t *testing.T) {
		var input controllers.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		// factory should have been called only twice total (once per session)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

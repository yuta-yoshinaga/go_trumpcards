package controllers_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"go_trumpcards/interface_adapters/controllers"
	"go_trumpcards/interface_adapters/controllers/usecases"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestBlackJackWebController_Method(t *testing.T) {
	bjiMock := new(usecases.MockBlackJackInteractor)
	bjiMock.On("Reset").Return(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`).Times(2)
	bjiMock.On("Hit").Return(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	bjiMock.On("Stand").Return(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	tbc := controllers.NewBlackJackWebController(bjiMock)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/blackjac/exec", tbc.Exec),
	)
	api.SetApp(router)
	var jsonCase1 controllers.BlackJackWebInput
	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec h", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "h"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec hit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "hit"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec s", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "s"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("success Exec stand", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "stand"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"bye."}`)
	})
	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"Unsupported command."}`)
	})
	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": ""}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"param error."}`)
	})
	t.Run("failed Exec response empty", func(t *testing.T) {
		bjiMock.On("Reset").Return(``)
		_ = json.Unmarshal([]byte(`{"command": "r"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjac/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"score":0,"cards":null},"player":{"score":0,"cards":null},"message":"error."}`)
	})
}

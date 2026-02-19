package controllers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers/usecases"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/mock"
)

func TestPokerWebController_Method(t *testing.T) {
	mockOutput := `{"dealer":{"handRank":0,"handName":"","cards":null},"player":{"handRank":0,"handName":"","cards":null},"phase":1,"message":""}`
	piMock := new(usecases.MockPokerInteractor)
	piMock.On("Reset").Return(mockOutput).Times(2)
	piMock.On("Exchange", mock.Anything).Return(mockOutput)
	piMock.On("Stand").Return(mockOutput)

	tpc := controllers.NewPokerWebController(piMock)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/poker/exec", tpc.Exec),
	)
	api.SetApp(router)

	var jsonInput controllers.PokerWebInput
	emptyBody := `{"dealer":{"handRank":0,"handName":"","cards":null},"player":{"handRank":0,"handName":"","cards":null},"phase":0,"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(emptyBody)
	})
	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(emptyBody)
	})
	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec e", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "e", "indices": [0, 1]}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec exchange no indices", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "exchange"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec s", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "s"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec stand", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "stand"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"handRank":0,"handName":"","cards":null},"player":{"handRank":0,"handName":"","cards":null},"phase":0,"message":"Unsupported command."}`)
	})
	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"handRank":0,"handName":"","cards":null},"player":{"handRank":0,"handName":"","cards":null},"phase":0,"message":"param error."}`)
	})
	t.Run("failed Exec response empty", func(t *testing.T) {
		piMock.On("Reset").Return(``)
		_ = json.Unmarshal([]byte(`{"command": "r"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/poker/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"dealer":{"handRank":0,"handName":"","cards":null},"player":{"handRank":0,"handName":"","cards":null},"phase":0,"message":"error."}`)
	})
}

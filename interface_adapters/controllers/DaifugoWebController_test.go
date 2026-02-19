package controllers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers/usecases"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestDaifugoWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"lastPlay":null,"isRevolution":false,"gameEndFlag":false,"message":""}`
	expectedBody := `{"players":[],"currentTurn":0,"lastPlay":[],"isRevolution":false,"gameEndFlag":false,"message":""}`
	diMock := new(usecases.MockDaifugoInteractor)
	diMock.On("Reset").Return(mockOutput).Times(2)
	diMock.On("Play", []int{0}).Return(mockOutput).Times(2)
	diMock.On("Pass").Return(mockOutput).Times(2)

	dwc := controllers.NewDaifugoWebController(diMock)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/daifugo/exec", dwc.Exec),
	)
	api.SetApp(router)

	var jsonInput controllers.DaifugoWebInput

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/daifugo/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"lastPlay":[],"isRevolution":false,"gameEndFlag":false,"message":"bye."}`)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/daifugo/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "play", "cardIndices": [0]}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/daifugo/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec pass", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "pass"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/daifugo/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
}

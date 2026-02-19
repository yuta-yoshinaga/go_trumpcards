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

func TestOldMaidWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDiscardedPairs":0,"hasDrawn":false,"message":""}`
	// After controller unmarshal+remarshal, new fields are included
	expectedBody := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":""}`
	omiMock := new(usecases.MockOldMaidInteractor)
	omiMock.On("Reset").Return(mockOutput).Times(2)
	omiMock.On("Draw", -1).Return(mockOutput)

	towc := controllers.NewOldMaidWebController(omiMock)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/oldmaid/exec", towc.Exec),
	)
	api.SetApp(router)

	var jsonInput controllers.OldMaidWebInput
	// When "q" / "quit": responseStr = {"message":"bye."} → all other fields default to zero
	qBody := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})
	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})
	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
	t.Run("success Exec d", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "d"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
	t.Run("success Exec draw", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "draw"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"Unsupported command."}`)
	})
	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"param error."}`)
	})
	t.Run("failed Exec response empty", func(t *testing.T) {
		omiMock.On("Reset").Return(``)
		_ = json.Unmarshal([]byte(`{"command": "r"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"error."}`)
	})
}

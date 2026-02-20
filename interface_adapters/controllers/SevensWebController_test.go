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

func TestSevensWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	expectedBody := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	sgiMock := new(usecases.MockSevensInteractor)
	sgiMock.On("Reset").Return(mockOutput).Times(2)
	sgiMock.On("Play", -1).Return(mockOutput) // pass
	sgiMock.On("Play", 0).Return(mockOutput)

	tswc := controllers.NewSevensWebController(sgiMock)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/sevens/exec", tswc.Exec),
	)
	api.SetApp(router)

	var jsonInput controllers.SevensWebInput
	// For "q"/"quit": responseStr = {"message":"bye."} → other fields get zero values
	qBody := `{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p pass (no index, defaults to 0)", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "p"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p with index", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "p", "index": 0}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"Unsupported command."}`)
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})

	t.Run("failed Exec response empty", func(t *testing.T) {
		sgiMock.On("Reset").Return(``)
		_ = json.Unmarshal([]byte(`{"command": "r"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"error."}`)
	})
}

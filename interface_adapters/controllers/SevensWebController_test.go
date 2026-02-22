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

func TestSevensWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	expectedBody := mockOutput
	sgiMock := new(usecases.MockSevensInteractor)
	sgiMock.On("Reset").Return(mockOutput).Times(2)
	sgiMock.On("Play", -1).Return(mockOutput) // pass
	sgiMock.On("Play", 0).Return(mockOutput)
	sgiMock.On("PlayJoker", 0, 1, 6).Return(mockOutput)

	factory := func() uc.SevensInteractorIF { return sgiMock }
	tswc := controllers.NewSevensWebController(factory)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/sevens/exec", tswc.Exec),
	)
	api.SetApp(router)

	var jsonInput controllers.SevensWebInput
	// For "q"/"quit": responseStr = {"message":"bye."} → other fields get zero values
	qBody := `{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p pass (no index, defaults to 0)", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "p", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p with index", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "p", "index": 0, "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec j joker command", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "j", "index": 0, "jokerTargetSuit": 1, "jokerTargetValue": 6, "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"Unsupported command."}`)
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})
	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controllers.SevensWebInput{
			Command:   "reset",
			SessionId: strings.Repeat("a", controllers.SessionMaxIDLen+1),
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})

	t.Run("failed Exec response empty", func(t *testing.T) {
		sgiMock.On("Reset").Return(``)
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"error."}`)
	})
}

func TestSevensWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	mockA := new(usecases.MockSevensInteractor)
	mockA.On("Reset").Return(mockOutput)
	mockB := new(usecases.MockSevensInteractor)
	mockB.On("Reset").Return(mockOutput)

	callCount := 0
	isoController := controllers.NewSevensWebController(func() uc.SevensInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/sevens/exec", isoController.Exec),
	)
	api.SetApp(router)

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controllers.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset")
		mockB.AssertNotCalled(t, "Reset")
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controllers.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset")
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controllers.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

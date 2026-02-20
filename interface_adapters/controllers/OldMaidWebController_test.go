package controllers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers/usecases"
	uc "github.com/yuta-yoshinaga/go_trumpcards/usecases"

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

	factory := func() uc.OldMaidInteractorIF { return omiMock }
	towc := controllers.NewOldMaidWebController(factory)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/oldmaid/exec", towc.Exec),
	)
	api.SetApp(router)

	var jsonInput controllers.OldMaidWebInput
	// When "q" / "quit": responseStr = {"message":"bye."} → all other fields default to zero
	qBody := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})
	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})
	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
	t.Run("success Exec d", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "d", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
	t.Run("success Exec draw", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "draw", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})
	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"Unsupported command."}`)
	})
	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"param error."}`)
	})
	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"param error."}`)
	})
	t.Run("failed Exec response empty", func(t *testing.T) {
		omiMock.On("Reset").Return(``)
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"hasDrawn":false,"cpuActions":[],"message":"error."}`)
	})
}

func TestOldMaidWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDiscardedPairs":0,"hasDrawn":false,"message":""}`
	mockA := new(usecases.MockOldMaidInteractor)
	mockA.On("Reset").Return(mockOutput)
	mockB := new(usecases.MockOldMaidInteractor)
	mockB.On("Reset").Return(mockOutput)

	callCount := 0
	isoController := controllers.NewOldMaidWebController(func() uc.OldMaidInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/oldmaid/exec", isoController.Exec),
	)
	api.SetApp(router)

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controllers.OldMaidWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset")
		mockB.AssertNotCalled(t, "Reset")
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controllers.OldMaidWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset")
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controllers.OldMaidWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

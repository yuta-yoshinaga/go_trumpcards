package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

func TestSevensWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":5},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	expectedBody := mockOutput
	sgiMock := new(usecase.MockSevensInteractor)
	sgiMock.On("Reset").Return(mockOutput).Times(2)
	sgiMock.On("Play", -1).Return(mockOutput) // pass
	sgiMock.On("Play", 0).Return(mockOutput)
	sgiMock.On("PlayJoker", 0, 1, 6).Return(mockOutput)

	factory := func() uc.SevensInteractorIF { return sgiMock }
	tswc := controller.NewSevensWebController(factory)
	defer tswc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/sevens/exec", tswc.Exec),
	)
	api.SetApp(router)

	var jsonInput controller.SevensWebInput
	// For "q"/"quit": responseStr = {"message":"bye."} → other fields get zero values
	qBody := `{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":0},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"bye."}`

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
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":0},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"Unsupported command."}`)
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":0},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":0},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})
	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.SevensWebInput{
			Command:   "reset",
			SessionId: strings.Repeat("a", controller.SessionMaxIDLen+1),
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":0},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})

}

func TestSevensWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":true,"jokerCount":2,"cpuStrategy":true,"maxPasses":5},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`

	t.Run("reset with all config fields calls ResetWithConfig", func(t *testing.T) {
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", true, 2, true, 5).Return(mockOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/sevens/exec", tswc.Exec))
		api.SetApp(router)

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "tunnelEnabled": true, "jokerCount": 2, "cpuStrategy": true, "sessionId": "test-cfg-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", true, 2, true, 5)
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset without config fields calls Reset", func(t *testing.T) {
		defaultOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":5},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("Reset").Return(defaultOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/sevens/exec", tswc.Exec))
		api.SetApp(router)

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-cfg-2"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(defaultOutput)
		sgiMock.AssertCalled(t, "Reset")
		sgiMock.AssertNotCalled(t, "ResetWithConfig")
	})

	t.Run("reset with partial config calls ResetWithConfig with defaults", func(t *testing.T) {
		partialOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":true,"jokerCount":0,"cpuStrategy":false,"maxPasses":5},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", true, 0, false, 5).Return(partialOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/sevens/exec", tswc.Exec))
		api.SetApp(router)

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "tunnelEnabled": true, "sessionId": "test-cfg-3"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(partialOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", true, 0, false, 5)
	})

	t.Run("reset with maxPasses field calls ResetWithConfig", func(t *testing.T) {
		passesOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":3},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", false, 0, false, 3).Return(passesOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/sevens/exec", tswc.Exec))
		api.SetApp(router)

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "maxPasses": 3, "sessionId": "test-cfg-4"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(passesOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", false, 0, false, 3)
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("failed reset with invalid jokerCount negative", func(t *testing.T) {
		sgiMock := new(usecase.MockSevensInteractor)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/sevens/exec", tswc.Exec))
		api.SetApp(router)

		negOne := -1
		input := controller.SevensWebInput{
			Command:    "reset",
			JokerCount: &negOne,
			SessionId:  "test-cfg-neg",
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":0},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error: jokerCount must be between 0 and 2."}`)
		sgiMock.AssertNotCalled(t, "ResetWithConfig")
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("failed reset with invalid jokerCount too large", func(t *testing.T) {
		sgiMock := new(usecase.MockSevensInteractor)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/sevens/exec", tswc.Exec))
		api.SetApp(router)

		hundred := 100
		input := controller.SevensWebInput{
			Command:    "reset",
			JokerCount: &hundred,
			SessionId:  "test-cfg-big",
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":0},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error: jokerCount must be between 0 and 2."}`)
		sgiMock.AssertNotCalled(t, "ResetWithConfig")
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset with only maxPasses field calls ResetWithConfig with default maxPasses", func(t *testing.T) {
		passesOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":0},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", false, 0, false, 0).Return(passesOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/sevens/exec", tswc.Exec))
		api.SetApp(router)

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "maxPasses": 0, "sessionId": "test-cfg-5"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(passesOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", false, 0, false, 0)
		sgiMock.AssertNotCalled(t, "Reset")
	})
}

func TestSevensWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"jokerCount":0,"cpuStrategy":false,"maxPasses":5},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	mockA := new(usecase.MockSevensInteractor)
	mockA.On("Reset").Return(mockOutput)
	mockB := new(usecase.MockSevensInteractor)
	mockB.On("Reset").Return(mockOutput)

	callCount := 0
	isoController := controller.NewSevensWebController(func() uc.SevensInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/sevens/exec", isoController.Exec),
	)
	api.SetApp(router)

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset")
		mockB.AssertNotCalled(t, "Reset")
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/sevens/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset")
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controller.SevensWebInput
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

func TestSevensWebController_Stop(t *testing.T) {
	sgiMock := new(usecase.MockSevensInteractor)
	factory := func() uc.SevensInteractorIF { return sgiMock }
	c := controller.NewSevensWebController(factory)
	c.Stop()
	c.Stop()
}

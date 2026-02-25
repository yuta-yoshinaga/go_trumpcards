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

func TestDoubtWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":""}`
	expectedBody := mockOutput

	dgiMock := new(usecase.MockDoubtInteractor)
	dgiMock.On("Reset").Return(mockOutput).Times(2)
	dgiMock.On("Play", []int{0}, 1).Return(mockOutput)
	dgiMock.On("Play", []int{0}, 13).Return(mockOutput)
	dgiMock.On("GetCpuDoubters").Return([]int{})
	dgiMock.On("ResolveDoubt", []int{0}).Return(mockOutput)
	dgiMock.On("ResolveDoubt", []int{}).Return(mockOutput)
	dgiMock.On("SkipDoubt").Return(mockOutput)

	factory := func() uc.DoubtInteractorIF { return dgiMock }
	tdwc := controller.NewDoubtWebController(factory)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/doubt/exec", tdwc.Exec),
	)
	api.SetApp(router)

	// For "q"/"quit": other fields get zero values
	qBody := `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec r", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p play", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "p", "cardIndices": [0], "claimedValue": 1, "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec d doubt (human doubts)", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "d", "doubterIndices": [0], "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec d doubt confirm (cpu only, no human)", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "d", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec s skip", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "s", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("failed Exec other", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "other", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"Unsupported command."}`)
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"param error."}`)
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"param error."}`)
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.DoubtWebInput{
			Command:   "reset",
			SessionId: strings.Repeat("a", controller.SessionMaxIDLen+1),
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"param error."}`)
	})

	t.Run("failed Exec response empty", func(t *testing.T) {
		dgiMock.On("Reset").Return(``)
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"error."}`)
	})

	claimedValueTests := []struct {
		name         string
		claimedValue int
		wantCode     int
		wantBody     string
	}{
		{
			name:         "success at max boundary (13)",
			claimedValue: 13,
			wantCode:     http.StatusOK,
			wantBody:     expectedBody,
		},
		{
			name:         "failed too low (0)",
			claimedValue: 0,
			wantCode:     http.StatusBadRequest,
			wantBody:     `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"param error: claimedValue must be between 1 and 13."}`,
		},
		{
			name:         "failed too high (14)",
			claimedValue: 14,
			wantCode:     http.StatusBadRequest,
			wantBody:     `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"param error: claimedValue must be between 1 and 13."}`,
		},
	}
	for _, tc := range claimedValueTests {
		t.Run("Exec p play claimedValue "+tc.name, func(t *testing.T) {
			input := controller.DoubtWebInput{
				Command:      "p",
				CardIndices:  []int{0},
				ClaimedValue: tc.claimedValue,
				SessionId:    "test-session-1",
			}
			req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
			req.Header.Set("Content-Type", "application/json;charset=UTF-8")
			recorded := test.RunRequest(t, api.MakeHandler(), req)
			recorded.CodeIs(tc.wantCode)
			recorded.ContentTypeIsJson()
			recorded.BodyIs(tc.wantBody)
		})
	}

	t.Run("failed Exec cardIndices too large", func(t *testing.T) {
		indices := make([]int, controller.MaxCardIndices+1)
		input := controller.DoubtWebInput{
			Command:     "p",
			CardIndices: indices,
			SessionId:   "test-session-1",
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"param error."}`)
	})
}

func TestDoubtWebController_SkipWithCpuDoubters(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":""}`

	dgiMock := new(usecase.MockDoubtInteractor)
	dgiMock.On("GetCpuDoubters").Return([]int{1, 2})
	dgiMock.On("ResolveDoubt", []int{1, 2}).Return(mockOutput)

	factory := func() uc.DoubtInteractorIF { return dgiMock }
	tdwc := controller.NewDoubtWebController(factory)

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/doubt/exec", tdwc.Exec),
	)
	api.SetApp(router)

	t.Run("skip with cpu doubters calls ResolveDoubt", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "s", "sessionId": "session-skip-doubters"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
		dgiMock.AssertCalled(t, "ResolveDoubt", []int{1, 2})
	})
}

func TestDoubtWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":""}`

	mockA := new(usecase.MockDoubtInteractor)
	mockA.On("Reset").Return(mockOutput)
	mockB := new(usecase.MockDoubtInteractor)
	mockB.On("Reset").Return(mockOutput)

	callCount := 0
	isoController := controller.NewDoubtWebController(func() uc.DoubtInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/doubt/exec", isoController.Exec),
	)
	api.SetApp(router)

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset")
		mockB.AssertNotCalled(t, "Reset")
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset")
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

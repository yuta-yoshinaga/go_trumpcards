package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
)

// mustDoubtOutputJSON constructs a default DoubtWebOutput with the given message
// and returns its JSON representation, mirroring newDefaultOutput in the controller.
func mustDoubtOutputJSON(msg string) string {
	out := &controller.DoubtWebOutput{
		Players:     []*controller.DoubtWebOutputPlayer{},
		CpuDoubters: []int{},
		CpuActions:  []*controller.DoubtWebOutputAction{},
		WinnerIdx:   -1,
		Message:     msg,
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustDoubtOutputJSON: %v", err))
	}
	return string(b)
}

func TestDoubtWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":""}`
	expectedBody := mockOutput

	dgiMock := new(usecase.MockDoubtInteractor)
	dgiMock.On("ResetWithConfig", domain.DefaultDoubtConfig()).Return(mockOutput)
	dgiMock.On("Play", []int{0}, 1).Return(mockOutput)
	dgiMock.On("Play", []int{0}, 13).Return(mockOutput)
	dgiMock.On("GetCpuDoubters").Return([]int{})
	dgiMock.On("ResolveDoubt", []int{0}).Return(mockOutput)
	dgiMock.On("ResolveDoubt", []int{}).Return(mockOutput)
	dgiMock.On("SkipDoubt").Return(mockOutput)

	factory := func() uc.DoubtInteractorIF { return dgiMock }
	tdwc := controller.NewDoubtWebController(factory)
	defer tdwc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/doubt/exec", tdwc.Exec),
	)
	api.SetApp(router)

	t.Run("success Exec q", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustDoubtOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustDoubtOutputJSON("bye."))
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
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustDoubtOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustDoubtOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var jsonInput controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustDoubtOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.DoubtWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustDoubtOutputJSON("param error."))
	})

	claimedValueErrBody := mustDoubtOutputJSON(fmt.Sprintf(
		"param error: claimedValue must be between %d and %d.",
		domain.MinClaimedValue, domain.MaxClaimedValue,
	))
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
			wantBody:     claimedValueErrBody,
		},
		{
			name:         "failed too high (14)",
			claimedValue: 14,
			wantCode:     http.StatusBadRequest,
			wantBody:     claimedValueErrBody,
		},
	}
	for _, tc := range claimedValueTests {
		t.Run("Exec p play claimedValue "+tc.name, func(t *testing.T) {
			input := controller.DoubtWebInput{
				BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
				CardIndices:  []int{0},
				ClaimedValue: tc.claimedValue,
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
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndices:  indices,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustDoubtOutputJSON("param error."))
	})
}

func TestDoubtWebController_SkipWithCpuDoubters(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":""}`

	dgiMock := new(usecase.MockDoubtInteractor)
	dgiMock.On("GetCpuDoubters").Return([]int{1, 2})
	dgiMock.On("ResolveDoubt", []int{1, 2}).Return(mockOutput)

	factory := func() uc.DoubtInteractorIF { return dgiMock }
	tdwc := controller.NewDoubtWebController(factory)
	defer tdwc.Stop()

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
	mockA.On("ResetWithConfig", domain.DefaultDoubtConfig()).Return(mockOutput)
	mockB := new(usecase.MockDoubtInteractor)
	mockB.On("ResetWithConfig", domain.DefaultDoubtConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewDoubtWebController(func() uc.DoubtInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

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
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultDoubtConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultDoubtConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.DoubtWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultDoubtConfig())
	})

	t.Run("session-A second call reuses mockA without creating new interactor", func(t *testing.T) {
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

func TestDoubtWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"phase":0,"tableCardCount":0,"lastAction":null,"cpuDoubters":[],"cpuActions":[],"humanAction":null,"lastDoubtResult":null,"gameEndFlag":false,"winnerIdx":-1,"message":"","doubtWindowSec":3}`

	t.Run("custom doubtWindowSec and cpuMemoryLevel are passed", func(t *testing.T) {
		win := 3
		mem := 2
		expected := domain.DoubtConfig{DoubtWindowSec: 3, CpuMemoryLevel: domain.DoubtMemoryLevelHard}
		dgiMock := new(usecase.MockDoubtInteractor)
		dgiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.DoubtInteractorIF { return dgiMock }
		ctrl := controller.NewDoubtWebController(factory)
		defer ctrl.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/doubt/exec", ctrl.Exec))
		api.SetApp(router)

		input := controller.DoubtWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			DoubtWindowSec: &win,
			CpuMemoryLevel: &mem,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		dgiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("doubtWindowSec below min (0) is ignored, uses default", func(t *testing.T) {
		win := 0
		mem := 1
		expected := domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelNormal}
		dgiMock := new(usecase.MockDoubtInteractor)
		dgiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.DoubtInteractorIF { return dgiMock }
		ctrl := controller.NewDoubtWebController(factory)
		defer ctrl.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/doubt/exec", ctrl.Exec))
		api.SetApp(router)

		input := controller.DoubtWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			DoubtWindowSec: &win,
			CpuMemoryLevel: &mem,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		dgiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuMemoryLevel above max (3) is ignored, uses default", func(t *testing.T) {
		win := 5
		mem := 3 // out of range [0-2]
		expected := domain.DoubtConfig{DoubtWindowSec: 5, CpuMemoryLevel: domain.DoubtMemoryLevelNormal}
		dgiMock := new(usecase.MockDoubtInteractor)
		dgiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.DoubtInteractorIF { return dgiMock }
		ctrl := controller.NewDoubtWebController(factory)
		defer ctrl.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/doubt/exec", ctrl.Exec))
		api.SetApp(router)

		input := controller.DoubtWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			DoubtWindowSec: &win,
			CpuMemoryLevel: &mem,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		dgiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuMemoryLevel below min (-1) is ignored, uses default", func(t *testing.T) {
		win := 5
		mem := -1 // below 0
		expected := domain.DoubtConfig{DoubtWindowSec: 5, CpuMemoryLevel: domain.DoubtMemoryLevelNormal}
		dgiMock := new(usecase.MockDoubtInteractor)
		dgiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.DoubtInteractorIF { return dgiMock }
		ctrl := controller.NewDoubtWebController(factory)
		defer ctrl.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/doubt/exec", ctrl.Exec))
		api.SetApp(router)

		input := controller.DoubtWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			DoubtWindowSec: &win,
			CpuMemoryLevel: &mem,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		dgiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("penaltyDrawLimit config values", func(t *testing.T) {
		win := 10
		testCases := []struct {
			name        string
			limit       *int
			expectedCfg domain.DoubtConfig
		}{
			{
				name:        "valid value is passed",
				limit:       func() *int { l := 5; return &l }(),
				expectedCfg: domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelNormal, PenaltyDrawLimit: 5},
			},
			{
				name:        "negative is ignored, uses default (0)",
				limit:       func() *int { l := -1; return &l }(),
				expectedCfg: domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelNormal},
			},
			{
				name:        "zero is valid (unlimited)",
				limit:       func() *int { l := 0; return &l }(),
				expectedCfg: domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelNormal, PenaltyDrawLimit: 0},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				dgiMock := new(usecase.MockDoubtInteractor)
				dgiMock.On("ResetWithConfig", tc.expectedCfg).Return(mockOutput)

				factory := func() uc.DoubtInteractorIF { return dgiMock }
				ctrl := controller.NewDoubtWebController(factory)
				defer ctrl.Stop()
				api := rest.NewApi()
				router, _ := rest.MakeRouter(rest.Post("/doubt/exec", ctrl.Exec))
				api.SetApp(router)

				input := controller.DoubtWebInput{
					BaseWebInput:     controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-pdl"},
					DoubtWindowSec:   &win,
					PenaltyDrawLimit: tc.limit,
				}
				req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
				req.Header.Set("Content-Type", "application/json;charset=UTF-8")
				recorded := test.RunRequest(t, api.MakeHandler(), req)
				recorded.CodeIs(http.StatusOK)
				dgiMock.AssertCalled(t, "ResetWithConfig", tc.expectedCfg)
			})
		}
	})
}

func TestDoubtWebController_ResetWithCpuHesitation(t *testing.T) {
	mockOutput := `{"players":[],"cpuDoubters":[],"cpuActions":[]}`
	win := 10

	t.Run("cpuHesitationEnabled true is passed", func(t *testing.T) {
		hesitation := true
		expected := domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelNormal, CpuHesitationEnabled: true}
		dgiMock := new(usecase.MockDoubtInteractor)
		dgiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.DoubtInteractorIF { return dgiMock }
		ctrl := controller.NewDoubtWebController(factory)
		defer ctrl.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/doubt/exec", ctrl.Exec))
		api.SetApp(router)

		input := controller.DoubtWebInput{
			BaseWebInput:         controller.BaseWebInput{Command: "reset", SessionID: "cfg-hes"},
			DoubtWindowSec:       &win,
			CpuHesitationEnabled: &hesitation,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		dgiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuHesitationEnabled nil uses default (false)", func(t *testing.T) {
		expected := domain.DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: domain.DoubtMemoryLevelNormal}
		dgiMock := new(usecase.MockDoubtInteractor)
		dgiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.DoubtInteractorIF { return dgiMock }
		ctrl := controller.NewDoubtWebController(factory)
		defer ctrl.Stop()
		api := rest.NewApi()
		router, _ := rest.MakeRouter(rest.Post("/doubt/exec", ctrl.Exec))
		api.SetApp(router)

		input := controller.DoubtWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "reset", SessionID: "cfg-hes-nil"},
			DoubtWindowSec: &win,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/doubt/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		dgiMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestDoubtWebController_Stop(t *testing.T) {
	dgiMock := new(usecase.MockDoubtInteractor)
	factory := func() uc.DoubtInteractorIF { return dgiMock }
	c := controller.NewDoubtWebController(factory)
	c.Stop()
	c.Stop()
}

func TestDoubtWebOutputAction_HasTell_JSON(t *testing.T) {
	action := &controller.DoubtWebOutputAction{
		PlayerIdx:    1,
		ClaimedValue: 5,
		CardCount:    2,
		IsBluff:      false,
		HasTell:      true,
	}
	b, err := json.Marshal(action)
	if err != nil {
		t.Fatal(err)
	}
	var parsed controller.DoubtWebOutputAction
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatal(err)
	}
	if !parsed.HasTell {
		t.Error("expected HasTell to be true after JSON round-trip")
	}
	if parsed.PlayerIdx != 1 || parsed.ClaimedValue != 5 || parsed.CardCount != 2 {
		t.Error("unexpected field values after round-trip")
	}

	// Also verify HasTell=false round-trip
	action.HasTell = false
	b, _ = json.Marshal(action)
	_ = json.Unmarshal(b, &parsed)
	if parsed.HasTell {
		t.Error("expected HasTell to be false after JSON round-trip")
	}
}

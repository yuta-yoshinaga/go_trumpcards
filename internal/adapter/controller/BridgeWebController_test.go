//go:build test

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
)

func mustBridgeOutputJSON(msg string) string {
	out := &controller.BridgeWebOutput{
		Players:       []*controller.BridgeWebOutputPlayer{},
		BidHistory:    []*controller.BridgeWebOutputBidEntry{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		DummyHand:     []*controller.WebOutputCard{},
		WinnerTeam:    -1,
		DeclarerIdx:   -1,
		DummyIdx:      -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBridgeOutputJSON: %v", err))
	}
	return string(b)
}

func TestBridgeWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"trickNumber":0,"currentPlayerIdx":0,"bidPlayerIdx":0,"dealerIdx":0,"trumpSuit":0,"contractLevel":0,"contractSuit":0,"doubled":0,"declarerIdx":0,"dummyIdx":0,"bidHistory":[],"vulnerability":[false,false],"currentTrick":[],"teamScores":[0,0],"gamesWon":[0,0],"belowLine":[0,0],"gameEndFlag":false,"winnerTeam":0,"leadPlayerIdx":0,"openingLeadDone":false,"dummyHand":[],"message":"","config":{"cpuDifficulty":0}}`
	expectedBody := mockOutput

	biMock := new(usecase.MockBridgeInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBridgeConfig()).Return(mockOutput)
	biMock.On("Bid", 0, 0, 0).Return(mockOutput)
	biMock.On("Bid", 1, 2, 3).Return(mockOutput)
	biMock.On("Play", 3).Return(mockOutput)
	biMock.On("NextTrick").Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BridgeInteractorIF { return biMock }
	ctrl := controller.NewBridgeWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBridgeOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBridgeOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec bid with no params (pass)", func(t *testing.T) {
		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "test-session-1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec b shorthand bid", func(t *testing.T) {
		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "test-session-1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec bid with all params", func(t *testing.T) {
		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "test-session-1"},
			BidType:      func() *int { v := 1; return &v }(),
			BidLevel:     func() *int { v := 2; return &v }(),
			BidSuit:      func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p play", func(t *testing.T) {
		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec n next", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec next", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"next","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec h hint", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec hint", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBridgeOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBridgeOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBridgeOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBridgeOutputJSON("param error."))
	})

	t.Run("failed Exec play no cardIndex", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustBridgeOutputJSON("param error: cardIndex is required."))
	})
}

func TestBridgeWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"bidHistory":[],"currentTrick":[],"dummyHand":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		expected := domain.BridgeConfig{CpuDifficulty: domain.BridgeCpuDifficultyHard}
		biMock := new(usecase.MockBridgeInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BridgeInteractorIF { return biMock }
		ctrl := controller.NewBridgeWebController(factory)
		defer ctrl.Stop()

		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.BridgeWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored, uses default", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultBridgeConfig()
		biMock := new(usecase.MockBridgeInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BridgeInteractorIF { return biMock }
		ctrl := controller.NewBridgeWebController(factory)
		defer ctrl.Stop()

		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.BridgeWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored, uses default", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultBridgeConfig()
		biMock := new(usecase.MockBridgeInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BridgeInteractorIF { return biMock }
		ctrl := controller.NewBridgeWebController(factory)
		defer ctrl.Stop()

		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.BridgeWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultBridgeConfig()
		biMock := new(usecase.MockBridgeInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BridgeInteractorIF { return biMock }
		ctrl := controller.NewBridgeWebController(factory)
		defer ctrl.Stop()

		input := controller.BridgeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestBridgeWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"bidHistory":[],"currentTrick":[],"dummyHand":[]}`

	mockA := new(usecase.MockBridgeInteractor)
	mockA.On("ResetWithConfig", domain.DefaultBridgeConfig()).Return(mockOutput)
	mockB := new(usecase.MockBridgeInteractor)
	mockB.On("ResetWithConfig", domain.DefaultBridgeConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewBridgeWebController(func() uc.BridgeInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultBridgeConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultBridgeConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultBridgeConfig())
	})

	t.Run("session-A second call reuses mockA without creating new interactor", func(t *testing.T) {
		var input controller.BridgeWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestBridgeWebController_Stop(t *testing.T) {
	biMock := new(usecase.MockBridgeInteractor)
	factory := func() uc.BridgeInteractorIF { return biMock }
	c := controller.NewBridgeWebController(factory)
	c.Stop()
	c.Stop()
}

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

func mustCrazyEightsOutputJSON(msg string) string {
	out := &controller.CrazyEightsWebOutput{
		Players:       []*controller.CrazyEightsWebOutputPlayer{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCrazyEightsOutputJSON: %v", err))
	}
	return string(b)
}

func TestCrazyEightsWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"currentPlayerIdx":0,"discardTop":null,"drawPileCount":0,"chosenSuit":0,"gameEndFlag":false,"winnerIdx":-1,"message":"","config":{"cpuDifficulty":0,"pointLimit":0}}`
	expectedBody := mockOutput

	siMock := new(usecase.MockCrazyEightsInteractor)
	siMock.On("ResetWithConfig", domain.DefaultCrazyEightsConfig()).Return(mockOutput)
	siMock.On("Play", 3).Return(mockOutput)
	siMock.On("ChooseSuit", 2).Return(mockOutput)
	siMock.On("Draw").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)

	factory := func() uc.CrazyEightsInteractorIF { return siMock }
	ctrl := controller.NewCrazyEightsWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCrazyEightsOutputJSON("bye."))
	})

	// The Web CLI sends "hint"/"h" here. Before #5791 the default branch used
	// dispatchLog, which matches neither, so the request 400'd with
	// "Unsupported command." despite the interactor having Hint() all along.
	for _, cmd := range []string{"h", "hint"} {
		t.Run("hint via "+cmd, func(t *testing.T) {
			var input controller.CrazyEightsWebInput
			_ = json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"test-session-1"}`), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		})
	}

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCrazyEightsOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec play", func(t *testing.T) {
		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p play shorthand", func(t *testing.T) {
		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec draw", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec draw long", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"draw","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec suit", func(t *testing.T) {
		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "suit", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec s suit shorthand", func(t *testing.T) {
		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "s", SessionID: "test-session-1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCrazyEightsOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCrazyEightsOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCrazyEightsOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCrazyEightsOutputJSON("param error."))
	})

	t.Run("failed Exec play no cardIndex", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCrazyEightsOutputJSON("param error: cardIndex is required."))
	})

	t.Run("failed Exec suit no suit", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"suit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustCrazyEightsOutputJSON("param error: suit is required."))
	})
}

func TestCrazyEightsWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 300
		expected := domain.CrazyEightsConfig{CpuDifficulty: domain.CrazyEightsCpuDifficultyHard, PointLimit: 300}
		siMock := new(usecase.MockCrazyEightsInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CrazyEightsInteractorIF { return siMock }
		ctrl := controller.NewCrazyEightsWebController(factory)
		defer ctrl.Stop()

		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.CrazyEightsWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored, uses default", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultCrazyEightsConfig()
		siMock := new(usecase.MockCrazyEightsInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CrazyEightsInteractorIF { return siMock }
		ctrl := controller.NewCrazyEightsWebController(factory)
		defer ctrl.Stop()

		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.CrazyEightsWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored, uses default", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultCrazyEightsConfig()
		siMock := new(usecase.MockCrazyEightsInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CrazyEightsInteractorIF { return siMock }
		ctrl := controller.NewCrazyEightsWebController(factory)
		defer ctrl.Stop()

		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.CrazyEightsWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored, uses default", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultCrazyEightsConfig()
		siMock := new(usecase.MockCrazyEightsInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CrazyEightsInteractorIF { return siMock }
		ctrl := controller.NewCrazyEightsWebController(factory)
		defer ctrl.Stop()

		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.CrazyEightsWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit exceeding 1000 is ignored", func(t *testing.T) {
		limit := 1001
		expected := domain.DefaultCrazyEightsConfig()
		siMock := new(usecase.MockCrazyEightsInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CrazyEightsInteractorIF { return siMock }
		ctrl := controller.NewCrazyEightsWebController(factory)
		defer ctrl.Stop()

		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-limit-max"},
			Config:       &controller.CrazyEightsWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultCrazyEightsConfig()
		siMock := new(usecase.MockCrazyEightsInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.CrazyEightsInteractorIF { return siMock }
		ctrl := controller.NewCrazyEightsWebController(factory)
		defer ctrl.Stop()

		input := controller.CrazyEightsWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestCrazyEightsWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[]}`

	mockA := new(usecase.MockCrazyEightsInteractor)
	mockA.On("ResetWithConfig", domain.DefaultCrazyEightsConfig()).Return(mockOutput)
	mockB := new(usecase.MockCrazyEightsInteractor)
	mockB.On("ResetWithConfig", domain.DefaultCrazyEightsConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewCrazyEightsWebController(func() uc.CrazyEightsInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultCrazyEightsConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultCrazyEightsConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultCrazyEightsConfig())
	})

	t.Run("session-A second call reuses mockA without creating new interactor", func(t *testing.T) {
		var input controller.CrazyEightsWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestCrazyEightsWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockCrazyEightsInteractor)
	factory := func() uc.CrazyEightsInteractorIF { return siMock }
	c := controller.NewCrazyEightsWebController(factory)
	c.Stop()
	c.Stop()
}

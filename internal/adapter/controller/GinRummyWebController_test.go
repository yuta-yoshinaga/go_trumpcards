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

func mustGinRummyOutputJSON(msg string) string {
	out := &controller.GinRummyWebOutput{
		Players:         []*controller.GinRummyWebOutputPlayer{},
		WinnerIdx:       -1,
		KnockerIdx:      -1,
		KnockerMelds:    []*controller.GinRummyWebOutputMeld{},
		KnockerDeadwood: []*controller.WebOutputCard{},
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGinRummyOutputJSON: %v", err))
	}
	return string(b)
}

func TestGinRummyWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"roundNumber":0,"currentPlayerIdx":0,"discardTop":null,"drawPileCount":0,"gameEndFlag":false,"winnerIdx":-1,"knockerIdx":-1,"knockerMelds":[],"knockerDeadwood":[],"isGin":false,"message":"","config":{"cpuDifficulty":0,"pointLimit":0}}`
	expectedBody := mockOutput

	siMock := new(usecase.MockGinRummyInteractor)
	siMock.On("ResetWithConfig", domain.DefaultGinRummyConfig()).Return(mockOutput)
	siMock.On("DrawFromStock").Return(mockOutput)
	siMock.On("DrawFromDiscard").Return(mockOutput)
	siMock.On("Discard", 3).Return(mockOutput)
	siMock.On("Knock", 3).Return(mockOutput)
	siMock.On("Layoff", []int{1, 2}).Return(mockOutput)
	siMock.On("Layoff", ([]int)(nil)).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.GinRummyInteractorIF { return siMock }
	ctrl := controller.NewGinRummyWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustGinRummyOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustGinRummyOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec ds drawstock", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"ds","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec drawstock", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"drawstock","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec dd drawdiscard", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"dd","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec drawdiscard", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"drawdiscard","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec discard", func(t *testing.T) {
		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec discard long", func(t *testing.T) {
		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "discard", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec knock", func(t *testing.T) {
		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "k", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec knock long", func(t *testing.T) {
		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "knock", SessionID: "test-session-1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec layoff", func(t *testing.T) {
		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "lo", SessionID: "test-session-1"},
			CardIndices:  []int{1, 2},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec layoff long", func(t *testing.T) {
		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "layoff", SessionID: "test-session-1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nr nextround", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec nextround", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustGinRummyOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustGinRummyOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustGinRummyOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustGinRummyOutputJSON("param error."))
	})

	t.Run("failed Exec discard no cardIndex", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustGinRummyOutputJSON("param error: cardIndex is required."))
	})

	t.Run("failed Exec knock no cardIndex", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"k","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustGinRummyOutputJSON("param error: cardIndex is required."))
	})
}

func TestGinRummyWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config values are passed", func(t *testing.T) {
		diff := 2
		limit := 200
		expected := domain.GinRummyConfig{CpuDifficulty: domain.GinRummyCpuDifficultyHard, PointLimit: 200}
		siMock := new(usecase.MockGinRummyInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.GinRummyInteractorIF { return siMock }
		ctrl := controller.NewGinRummyWebController(factory)
		defer ctrl.Stop()

		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.GinRummyWebConfig{CpuDifficulty: &diff, PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultGinRummyConfig()
		siMock := new(usecase.MockGinRummyInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.GinRummyInteractorIF { return siMock }
		ctrl := controller.NewGinRummyWebController(factory)
		defer ctrl.Stop()

		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.GinRummyWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultGinRummyConfig()
		siMock := new(usecase.MockGinRummyInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.GinRummyInteractorIF { return siMock }
		ctrl := controller.NewGinRummyWebController(factory)
		defer ctrl.Stop()

		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.GinRummyWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit below 1 is ignored", func(t *testing.T) {
		limit := 0
		expected := domain.DefaultGinRummyConfig()
		siMock := new(usecase.MockGinRummyInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.GinRummyInteractorIF { return siMock }
		ctrl := controller.NewGinRummyWebController(factory)
		defer ctrl.Stop()

		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
			Config:       &controller.GinRummyWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("pointLimit exceeding 1000 is ignored", func(t *testing.T) {
		limit := 1001
		expected := domain.DefaultGinRummyConfig()
		siMock := new(usecase.MockGinRummyInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.GinRummyInteractorIF { return siMock }
		ctrl := controller.NewGinRummyWebController(factory)
		defer ctrl.Stop()

		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-limit-max"},
			Config:       &controller.GinRummyWebConfig{PointLimit: &limit},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultGinRummyConfig()
		siMock := new(usecase.MockGinRummyInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.GinRummyInteractorIF { return siMock }
		ctrl := controller.NewGinRummyWebController(factory)
		defer ctrl.Stop()

		input := controller.GinRummyWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-5"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestGinRummyWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[]}`

	mockA := new(usecase.MockGinRummyInteractor)
	mockA.On("ResetWithConfig", domain.DefaultGinRummyConfig()).Return(mockOutput)
	mockB := new(usecase.MockGinRummyInteractor)
	mockB.On("ResetWithConfig", domain.DefaultGinRummyConfig()).Return(mockOutput)

	callCount := 0
	isoController := controller.NewGinRummyWebController(func() uc.GinRummyInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "ResetWithConfig", domain.DefaultGinRummyConfig())
		mockB.AssertNotCalled(t, "ResetWithConfig", domain.DefaultGinRummyConfig())
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "ResetWithConfig", domain.DefaultGinRummyConfig())
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controller.GinRummyWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestGinRummyWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockGinRummyInteractor)
	factory := func() uc.GinRummyInteractorIF { return siMock }
	c := controller.NewGinRummyWebController(factory)
	c.Stop()
	c.Stop()
}

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

func mustMemoryOutputJSON(msg string) string {
	out := &controller.MemoryWebOutput{
		Players:       []*controller.MemoryWebOutputPlayer{},
		Board:         []*controller.MemoryWebOutputBoardCard{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMemoryOutputJSON: %v", err))
	}
	return string(b)
}

func TestMemoryWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"board":[],"phase":0,"currentPlayerIdx":0,"firstFlipPos":-1,"secondFlipPos":-1,"lastMatchResult":false,"gameEndFlag":false,"winnerIdx":-1,"turnNumber":0,"message":"","config":{"cpuDifficulty":0}}`
	expectedBody := mockOutput

	miMock := new(usecase.MockMemoryInteractor)
	miMock.On("ResetWithConfig", domain.DefaultMemoryConfig()).Return(mockOutput)
	miMock.On("Flip", 5).Return(mockOutput)
	miMock.On("Next").Return(mockOutput)
	miMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.MemoryInteractorIF { return miMock }
	ctrl := controller.NewMemoryWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustMemoryOutputJSON("bye."))
	})

	t.Run("success Exec quit", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustMemoryOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec f flip", func(t *testing.T) {
		input := controller.MemoryWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "f", SessionID: "test-session-1"},
			Position:     func() *int { v := 5; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec flip", func(t *testing.T) {
		input := controller.MemoryWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "flip", SessionID: "test-session-1"},
			Position:     func() *int { v := 5; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec n next", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec next", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"next","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec l shorthand", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("failed Exec other", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustMemoryOutputJSON("Unsupported command."))
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustMemoryOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustMemoryOutputJSON("param error."))
	})

	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.MemoryWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustMemoryOutputJSON("param error."))
	})

	t.Run("failed Exec flip no position", func(t *testing.T) {
		var input controller.MemoryWebInput
		_ = json.Unmarshal([]byte(`{"command":"f","sessionId":"test-session-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustMemoryOutputJSON("param error: position is required."))
	})
}

func TestMemoryWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"board":[]}`

	t.Run("custom cpuDifficulty is passed", func(t *testing.T) {
		diff := 2
		expected := domain.MemoryConfig{CpuDifficulty: domain.MemoryCpuDifficultyHard, PairCount: domain.MemoryMaxPairCount}
		miMock := new(usecase.MockMemoryInteractor)
		miMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.MemoryInteractorIF { return miMock }
		ctrl := controller.NewMemoryWebController(factory)
		defer ctrl.Stop()

		input := controller.MemoryWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-1"},
			Config:       &controller.MemoryWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		miMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty above max is ignored", func(t *testing.T) {
		diff := 3
		expected := domain.DefaultMemoryConfig()
		miMock := new(usecase.MockMemoryInteractor)
		miMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.MemoryInteractorIF { return miMock }
		ctrl := controller.NewMemoryWebController(factory)
		defer ctrl.Stop()

		input := controller.MemoryWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-2"},
			Config:       &controller.MemoryWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		miMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("cpuDifficulty below min is ignored", func(t *testing.T) {
		diff := -1
		expected := domain.DefaultMemoryConfig()
		miMock := new(usecase.MockMemoryInteractor)
		miMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.MemoryInteractorIF { return miMock }
		ctrl := controller.NewMemoryWebController(factory)
		defer ctrl.Stop()

		input := controller.MemoryWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-3"},
			Config:       &controller.MemoryWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		miMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultMemoryConfig()
		miMock := new(usecase.MockMemoryInteractor)
		miMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.MemoryInteractorIF { return miMock }
		ctrl := controller.NewMemoryWebController(factory)
		defer ctrl.Stop()

		input := controller.MemoryWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-session-4"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		miMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestMemoryWebController_Stop(t *testing.T) {
	miMock := new(usecase.MockMemoryInteractor)
	factory := func() uc.MemoryInteractorIF { return miMock }
	c := controller.NewMemoryWebController(factory)
	c.Stop()
	c.Stop()
}

//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustLooOutputJSON(msg string) string {
	out := &controller.LooWebOutput{
		Players:         []*controller.LooWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		LastTrick:       []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		TotalTricks:     domain.LooTrickCount,
		LastTrickWinner: -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustLooOutputJSON: %v", err))
	}
	return string(b)
}

func TestLooWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	liMock := new(usecase.MockLooInteractor)
	liMock.On("ResetWithConfig", domain.DefaultLooConfig()).Return(mockOutput)
	liMock.On("Decide", true).Return(mockOutput)
	liMock.On("Play", 3).Return(mockOutput)
	liMock.On("NextRound").Return(mockOutput)
	liMock.On("Hint").Return(mockOutput)
	liMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.LooInteractorIF { return liMock }
	ctrl := controller.NewLooWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.LooWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustLooOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("decide", func(t *testing.T) {
		input := controller.LooWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			Play:         func() *bool { v := true; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("decide missing", func(t *testing.T) {
		run(t, `{"command":"d","sessionId":"s1"}`, mustLooOutputJSON("param error: play is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.LooWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustLooOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("next alias", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustLooOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustLooOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestLooWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		ante := 5
		expected := domain.LooConfig{CpuDifficulty: domain.LooCpuDifficultyHard, Ante: 5}
		liMock := new(usecase.MockLooInteractor)
		liMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewLooWebController(func() uc.LooInteractorIF { return liMock })
		defer ctrl.Stop()

		input := controller.LooWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.LooWebConfig{CpuDifficulty: &diff, Ante: &ante},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		liMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultLooConfig()
		liMock := new(usecase.MockLooInteractor)
		liMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewLooWebController(func() uc.LooInteractorIF { return liMock })
		defer ctrl.Stop()

		input := controller.LooWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.LooWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		liMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultLooConfig()
		liMock := new(usecase.MockLooInteractor)
		liMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewLooWebController(func() uc.LooInteractorIF { return liMock })
		defer ctrl.Stop()

		input := controller.LooWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		liMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestLooWebController_Stop(t *testing.T) {
	liMock := new(usecase.MockLooInteractor)
	c := controller.NewLooWebController(func() uc.LooInteractorIF { return liMock })
	c.Stop()
	c.Stop()
}

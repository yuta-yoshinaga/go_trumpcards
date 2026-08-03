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

func mustViraOutputJSON(msg string) string {
	out := &controller.ViraWebOutput{
		Players:         []*controller.ViraWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		DeclarerIdx:     -1,
		WinnerPlayer:    -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustViraOutputJSON: %v", err))
	}
	return string(b)
}

func TestViraWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockViraInteractor)
	diMock.On("ResetWithConfig", domain.DefaultViraConfig()).Return(mockOutput)
	diMock.On("Bid", 1).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ViraInteractorIF { return diMock }
	ctrl := controller.NewViraWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.ViraWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustViraOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid", func(t *testing.T) {
		input := controller.ViraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Bid:          func() *int { v := 1; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid missing", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`, mustViraOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.ViraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustViraOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("next", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustViraOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustViraOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestViraWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	newCtrl := func(expected domain.ViraConfig) (*usecase.MockViraInteractor, *controller.ViraWebController) {
		diMock := new(usecase.MockViraInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		return diMock, controller.NewViraWebController(func() uc.ViraInteractorIF { return diMock })
	}

	t.Run("custom config passed through", func(t *testing.T) {
		diff, rounds := 2, 9
		expected := domain.ViraConfig{CpuDifficulty: domain.ViraCpuDifficultyHard, TargetRounds: 9}
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.ViraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.ViraWebConfig{CpuDifficulty: &diff, TargetRounds: &rounds},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultViraConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.ViraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.ViraWebConfig{CpuDifficulty: &diff},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	// A round count below the player count passes the bounds check but fails
	// ViraConfig.Validate, so the reset would be silently swallowed. Clamping to
	// the default here keeps the request honest.
	t.Run("rounds below the player count fall back to default", func(t *testing.T) {
		rounds := 1
		expected := domain.DefaultViraConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.ViraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c4"},
			Config:       &controller.ViraWebConfig{TargetRounds: &rounds},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultViraConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.ViraWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestViraWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockViraInteractor)
	c := controller.NewViraWebController(func() uc.ViraInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

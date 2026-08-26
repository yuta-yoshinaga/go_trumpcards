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

func mustTrappolaOutputJSON(msg string) string {
	out := &controller.TrappolaWebOutput{
		Players:         []*controller.TrappolaWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		LastTrick:       []*controller.WebOutputTrickCard{},
		LastTrickWinner: -1,
		TeamScores:      []int{},
		TeamRoundThirds: []int{},
		PlayableIndices: []int{},
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTrappolaOutputJSON: %v", err))
	}
	return string(b)
}

func TestTrappolaWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockTrappolaInteractor)
	giMock.On("ResetWithConfig", domain.DefaultTrappolaConfig()).Return(mockOutput)
	giMock.On("Play", 3).Return(mockOutput)
	giMock.On("NextTrick").Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TrappolaInteractorIF { return giMock }
	ctrl := controller.NewTrappolaWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		var input controller.TrappolaWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTrappolaOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.TrappolaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustTrappolaOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustTrappolaOutputJSON("param error."), http.StatusBadRequest)
	})
	t.Run("play no cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustTrappolaOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
}

func TestTrappolaWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		target := 31
		expected := domain.TrappolaConfig{CpuDifficulty: domain.TrappolaCpuDifficultyHard, TargetPoints: 31}
		giMock := new(usecase.MockTrappolaInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTrappolaWebController(func() uc.TrappolaInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TrappolaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.TrappolaWebConfig{CpuDifficulty: &diff, TargetPoints: &target},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		target := 21
		expected := domain.DefaultTrappolaConfig()
		giMock := new(usecase.MockTrappolaInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTrappolaWebController(func() uc.TrappolaInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TrappolaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.TrappolaWebConfig{CpuDifficulty: &diff, TargetPoints: &target},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultTrappolaConfig()
		giMock := new(usecase.MockTrappolaInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTrappolaWebController(func() uc.TrappolaInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TrappolaWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestTrappolaWebController_Stop(t *testing.T) {
	giMock := new(usecase.MockTrappolaInteractor)
	c := controller.NewTrappolaWebController(func() uc.TrappolaInteractorIF { return giMock })
	c.Stop()
	c.Stop()
}

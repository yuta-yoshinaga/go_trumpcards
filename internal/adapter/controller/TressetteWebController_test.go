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

func mustTressetteOutputJSON(msg string) string {
	out := &controller.TressetteWebOutput{
		Players:         []*controller.TressetteWebOutputPlayer{},
		CurrentTrick:    []*controller.TressetteWebOutputTrickCard{},
		LastTrick:       []*controller.TressetteWebOutputTrickCard{},
		LastTrickWinner: -1,
		TeamScores:      []int{},
		TeamRoundThirds: []int{},
		PlayableIndices: []int{},
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTressetteOutputJSON: %v", err))
	}
	return string(b)
}

func TestTressetteWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockTressetteInteractor)
	giMock.On("ResetWithConfig", domain.DefaultTressetteConfig()).Return(mockOutput)
	giMock.On("Play", 3).Return(mockOutput)
	giMock.On("NextTrick").Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TressetteInteractorIF { return giMock }
	ctrl := controller.NewTressetteWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		var input controller.TressetteWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTressetteOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.TressetteWebInput{
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustTressetteOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustTressetteOutputJSON("param error."), http.StatusBadRequest)
	})
	t.Run("play no cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustTressetteOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
}

func TestTressetteWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		target := 31
		expected := domain.TressetteConfig{CpuDifficulty: domain.TressetteCpuDifficultyHard, TargetPoints: 31}
		giMock := new(usecase.MockTressetteInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTressetteWebController(func() uc.TressetteInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TressetteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.TressetteWebConfig{CpuDifficulty: &diff, TargetPoints: &target},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		target := 21
		expected := domain.DefaultTressetteConfig()
		giMock := new(usecase.MockTressetteInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTressetteWebController(func() uc.TressetteInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TressetteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.TressetteWebConfig{CpuDifficulty: &diff, TargetPoints: &target},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultTressetteConfig()
		giMock := new(usecase.MockTressetteInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTressetteWebController(func() uc.TressetteInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.TressetteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestTressetteWebController_Stop(t *testing.T) {
	giMock := new(usecase.MockTressetteInteractor)
	c := controller.NewTressetteWebController(func() uc.TressetteInteractorIF { return giMock })
	c.Stop()
	c.Stop()
}

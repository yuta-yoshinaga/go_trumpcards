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

func mustMariasOutputJSON(msg string) string {
	out := &controller.MariasWebOutput{
		Players:         []*controller.MariasWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMariasOutputJSON: %v", err))
	}
	return string(b)
}

func TestMariasWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockMariasInteractor)
	diMock.On("ResetWithConfig", domain.DefaultMariasConfig()).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.MariasInteractorIF { return diMock }
	ctrl := controller.NewMariasWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.MariasWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustMariasOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.MariasWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustMariasOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustMariasOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustMariasOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestMariasWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		pts := 15
		expected := domain.MariasConfig{
			CpuDifficulty: domain.MariasCpuDifficultyHard,
			TargetPoints:  15,
		}
		diMock := new(usecase.MockMariasInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewMariasWebController(func() uc.MariasInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.MariasWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.MariasWebConfig{CpuDifficulty: &diff, TargetPoints: &pts},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultMariasConfig()
		diMock := new(usecase.MockMariasInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewMariasWebController(func() uc.MariasInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.MariasWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.MariasWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultMariasConfig()
		diMock := new(usecase.MockMariasInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewMariasWebController(func() uc.MariasInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.MariasWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestMariasWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockMariasInteractor)
	c := controller.NewMariasWebController(func() uc.MariasInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

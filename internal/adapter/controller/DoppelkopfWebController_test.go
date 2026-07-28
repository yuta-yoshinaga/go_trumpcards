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

func mustDoppelkopfOutputJSON(msg string) string {
	out := &controller.DoppelkopfWebOutput{
		Players:         []*controller.DoppelkopfWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		ReTeam:          make([]bool, domain.DoppelkopfPlayerCnt),
		PlayableIndices: []int{},
		WinnerIdx:       -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustDoppelkopfOutputJSON: %v", err))
	}
	return string(b)
}

func TestDoppelkopfWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockDoppelkopfInteractor)
	diMock.On("ResetWithConfig", domain.DefaultDoppelkopfConfig()).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("Announce").Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.DoppelkopfInteractorIF { return diMock }
	ctrl := controller.NewDoppelkopfWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.DoppelkopfWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustDoppelkopfOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.DoppelkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustDoppelkopfOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("announce", func(t *testing.T) {
		run(t, `{"command":"a","sessionId":"s1"}`, mockOutput, http.StatusOK)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustDoppelkopfOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustDoppelkopfOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestDoppelkopfWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		base := 3
		expected := domain.DoppelkopfConfig{
			CpuDifficulty: domain.DoppelkopfCpuDifficultyHard,
			BaseChips:     3,
			StartChips:    20,
			TargetChips:   40,
		}
		diMock := new(usecase.MockDoppelkopfInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewDoppelkopfWebController(func() uc.DoppelkopfInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.DoppelkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.DoppelkopfWebConfig{CpuDifficulty: &diff, BaseChips: &base},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultDoppelkopfConfig()
		diMock := new(usecase.MockDoppelkopfInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewDoppelkopfWebController(func() uc.DoppelkopfInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.DoppelkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.DoppelkopfWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultDoppelkopfConfig()
		diMock := new(usecase.MockDoppelkopfInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewDoppelkopfWebController(func() uc.DoppelkopfInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.DoppelkopfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestDoppelkopfWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockDoppelkopfInteractor)
	c := controller.NewDoppelkopfWebController(func() uc.DoppelkopfInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

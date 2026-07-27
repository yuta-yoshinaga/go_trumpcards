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

func mustTuteOutputJSON(msg string) string {
	out := &controller.TuteWebOutput{
		Players:         []*controller.TuteWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		DeclaredSuits:   make([]bool, domain.CardDesignMax+1),
		PlayableIndices: []int{},
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTuteOutputJSON: %v", err))
	}
	return string(b)
}

func TestTuteWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockTuteInteractor)
	diMock.On("ResetWithConfig", domain.DefaultTuteConfig()).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("DeclareMarriage", 2).Return(mockOutput)
	diMock.On("DeclareTute").Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TuteInteractorIF { return diMock }
	ctrl := controller.NewTuteWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.TuteWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTuteOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.TuteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustTuteOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("marriage", func(t *testing.T) {
		input := controller.TuteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			Suit:         func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("marriage missing suit", func(t *testing.T) {
		run(t, `{"command":"m","sessionId":"s1"}`, mustTuteOutputJSON("param error: suit is required."), http.StatusBadRequest)
	})
	t.Run("tute", func(t *testing.T) {
		run(t, `{"command":"tute","sessionId":"s1"}`, mockOutput, http.StatusOK)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustTuteOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustTuteOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestTuteWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		pts := 200
		expected := domain.TuteConfig{
			CpuDifficulty: domain.TuteCpuDifficultyHard,
			TargetPoints:  200,
		}
		diMock := new(usecase.MockTuteInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTuteWebController(func() uc.TuteInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.TuteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.TuteWebConfig{CpuDifficulty: &diff, TargetPoints: &pts},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultTuteConfig()
		diMock := new(usecase.MockTuteInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTuteWebController(func() uc.TuteInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.TuteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.TuteWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultTuteConfig()
		diMock := new(usecase.MockTuteInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTuteWebController(func() uc.TuteInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.TuteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestTuteWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockTuteInteractor)
	c := controller.NewTuteWebController(func() uc.TuteInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

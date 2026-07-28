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

func mustSheepsheadOutputJSON(msg string) string {
	out := &controller.SheepsheadWebOutput{
		Players:         []*controller.SheepsheadWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		Buried:          []*controller.WebOutputCard{},
		CallableSuits:   []int{},
		PlayableIndices: []int{},
		PickerIdx:       -1,
		PartnerIdx:      -1,
		WinnerIdx:       -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSheepsheadOutputJSON: %v", err))
	}
	return string(b)
}

func TestSheepsheadWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	giMock := new(usecase.MockSheepsheadInteractor)
	giMock.On("ResetWithConfig", domain.DefaultSheepsheadConfig()).Return(mockOutput)
	giMock.On("Pick", true).Return(mockOutput)
	giMock.On("Pick", false).Return(mockOutput)
	giMock.On("Bury", []int{0, 1}).Return(mockOutput)
	giMock.On("Call", 2).Return(mockOutput)
	giMock.On("Play", 3).Return(mockOutput)
	giMock.On("NextTrick").Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.SheepsheadInteractorIF { return giMock }
	ctrl := controller.NewSheepsheadWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.SheepsheadWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustSheepsheadOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("pick true", func(t *testing.T) {
		pickVal := true
		input := controller.SheepsheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "pick", SessionID: "s1"},
			Pick:         &pickVal,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("pick missing param", func(t *testing.T) {
		run(t, `{"command":"pick","sessionId":"s1"}`, mustSheepsheadOutputJSON("param error: pick is required."), http.StatusBadRequest)
	})
	t.Run("bury", func(t *testing.T) {
		input := controller.SheepsheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bury", SessionID: "s1"},
			BuryIndices:  []int{0, 1},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bury missing param", func(t *testing.T) {
		run(t, `{"command":"bury","sessionId":"s1"}`, mustSheepsheadOutputJSON("param error: buryIndices is required."), http.StatusBadRequest)
	})
	t.Run("call", func(t *testing.T) {
		suit := 2
		input := controller.SheepsheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "call", SessionID: "s1"},
			CallSuit:     &suit,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("call missing param", func(t *testing.T) {
		run(t, `{"command":"call","sessionId":"s1"}`, mustSheepsheadOutputJSON("param error: callSuit is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.SheepsheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustSheepsheadOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustSheepsheadOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustSheepsheadOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestSheepsheadWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		base := 3
		expected := domain.SheepsheadConfig{
			CpuDifficulty: domain.SheepsheadCpuDifficultyHard,
			BaseChips:     3,
			StartChips:    20,
			TargetChips:   40,
		}
		giMock := new(usecase.MockSheepsheadInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewSheepsheadWebController(func() uc.SheepsheadInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.SheepsheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.SheepsheadWebConfig{CpuDifficulty: &diff, BaseChips: &base},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultSheepsheadConfig()
		giMock := new(usecase.MockSheepsheadInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewSheepsheadWebController(func() uc.SheepsheadInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.SheepsheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.SheepsheadWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultSheepsheadConfig()
		giMock := new(usecase.MockSheepsheadInteractor)
		giMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewSheepsheadWebController(func() uc.SheepsheadInteractorIF { return giMock })
		defer ctrl.Stop()

		input := controller.SheepsheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		giMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestSheepsheadWebController_Stop(t *testing.T) {
	giMock := new(usecase.MockSheepsheadInteractor)
	c := controller.NewSheepsheadWebController(func() uc.SheepsheadInteractorIF { return giMock })
	c.Stop()
	c.Stop()
}

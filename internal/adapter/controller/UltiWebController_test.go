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

func mustUltiOutputJSON(msg string) string {
	out := &controller.UltiWebOutput{
		Players:         []*controller.UltiWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustUltiOutputJSON: %v", err))
	}
	return string(b)
}

func TestUltiWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockUltiInteractor)
	diMock.On("ResetWithConfig", domain.DefaultUltiConfig()).Return(mockOutput)
	diMock.On("Bid", domain.UltiContractParty, domain.CardDesignHeart).Return(mockOutput)
	diMock.On("Discard", []int{0, 1}).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.UltiInteractorIF { return diMock }
	ctrl := controller.NewUltiWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.UltiWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustUltiOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid", func(t *testing.T) {
		input := controller.UltiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Contract:     func() *string { v := "party"; return &v }(),
			TrumpSuit:    func() *int { v := domain.CardDesignHeart; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid missing contract", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`, mustUltiOutputJSON("param error: contract is required."), http.StatusBadRequest)
	})
	t.Run("discard", func(t *testing.T) {
		input := controller.UltiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			CardIndices:  []int{0, 1},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("discard missing indices", func(t *testing.T) {
		run(t, `{"command":"d","sessionId":"s1"}`, mustUltiOutputJSON("param error: cardIndices is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.UltiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustUltiOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustUltiOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustUltiOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestUltiWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		rounds := 7
		expected := domain.UltiConfig{CpuDifficulty: domain.UltiCpuDifficultyHard, TargetRounds: 7}
		diMock := new(usecase.MockUltiInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewUltiWebController(func() uc.UltiInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.UltiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.UltiWebConfig{CpuDifficulty: &diff, TargetRounds: &rounds},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultUltiConfig()
		diMock := new(usecase.MockUltiInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewUltiWebController(func() uc.UltiInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.UltiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.UltiWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultUltiConfig()
		diMock := new(usecase.MockUltiInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewUltiWebController(func() uc.UltiInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.UltiWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestUltiWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockUltiInteractor)
	c := controller.NewUltiWebController(func() uc.UltiInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

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

func mustOmbreOutputJSON(msg string) string {
	out := &controller.OmbreWebOutput{
		Players:         []*controller.OmbreWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		OmbreIdx:        -1,
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustOmbreOutputJSON: %v", err))
	}
	return string(b)
}

func TestOmbreWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockOmbreInteractor)
	diMock.On("ResetWithConfig", domain.DefaultOmbreConfig()).Return(mockOutput)
	diMock.On("Bid", domain.OmbreBidEntrar, domain.CardDesignHeart).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.OmbreInteractorIF { return diMock }
	ctrl := controller.NewOmbreWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.OmbreWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustOmbreOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid", func(t *testing.T) {
		input := controller.OmbreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Bid:          func() *int { v := int(domain.OmbreBidEntrar); return &v }(),
			TrumpSuit:    func() *int { v := domain.CardDesignHeart; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid missing bid", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`, mustOmbreOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.OmbreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustOmbreOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustOmbreOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustOmbreOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestOmbreWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		rounds := 7
		expected := domain.OmbreConfig{CpuDifficulty: domain.OmbreCpuDifficultyHard, TargetRounds: 7}
		diMock := new(usecase.MockOmbreInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewOmbreWebController(func() uc.OmbreInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.OmbreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.OmbreWebConfig{CpuDifficulty: &diff, TargetRounds: &rounds},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultOmbreConfig()
		diMock := new(usecase.MockOmbreInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewOmbreWebController(func() uc.OmbreInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.OmbreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.OmbreWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultOmbreConfig()
		diMock := new(usecase.MockOmbreInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewOmbreWebController(func() uc.OmbreInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.OmbreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestOmbreWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockOmbreInteractor)
	c := controller.NewOmbreWebController(func() uc.OmbreInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

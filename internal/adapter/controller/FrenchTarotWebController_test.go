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

func mustFrenchTarotOutputJSON(msg string) string {
	out := &controller.FrenchTarotWebOutput{
		Players:         []*controller.FrenchTarotWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		Chien:           []*controller.WebOutputCard{},
		PlayableIndices: []int{},
		DeclarerIdx:     -1,
		HighestBidder:   -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustFrenchTarotOutputJSON: %v", err))
	}
	return string(b)
}

func TestFrenchTarotWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockFrenchTarotInteractor)
	diMock.On("ResetWithConfig", domain.DefaultFrenchTarotConfig()).Return(mockOutput)
	diMock.On("Bid", domain.FrenchTarotBidGarde).Return(mockOutput)
	diMock.On("Pass").Return(mockOutput)
	diMock.On("Discard", []int{0, 1, 2, 3, 4, 5}).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.FrenchTarotInteractorIF { return diMock }
	ctrl := controller.NewFrenchTarotWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.FrenchTarotWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustFrenchTarotOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid", func(t *testing.T) {
		input := controller.FrenchTarotWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Bid:          func() *string { v := "garde"; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid missing bid", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`, mustFrenchTarotOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("pass", func(t *testing.T) {
		run(t, `{"command":"pass","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("discard", func(t *testing.T) {
		input := controller.FrenchTarotWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			CardIndices:  []int{0, 1, 2, 3, 4, 5},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("discard missing indices", func(t *testing.T) {
		run(t, `{"command":"d","sessionId":"s1"}`, mustFrenchTarotOutputJSON("param error: cardIndices is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.FrenchTarotWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustFrenchTarotOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustFrenchTarotOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
}

func TestFrenchTarotWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		deals := 7
		expected := domain.FrenchTarotConfig{CpuDifficulty: domain.FrenchTarotCpuDifficultyHard, TargetDeals: 7}
		diMock := new(usecase.MockFrenchTarotInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewFrenchTarotWebController(func() uc.FrenchTarotInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.FrenchTarotWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.FrenchTarotWebConfig{CpuDifficulty: &diff, TargetDeals: &deals},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultFrenchTarotConfig()
		diMock := new(usecase.MockFrenchTarotInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewFrenchTarotWebController(func() uc.FrenchTarotInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.FrenchTarotWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestFrenchTarotWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockFrenchTarotInteractor)
	c := controller.NewFrenchTarotWebController(func() uc.FrenchTarotInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

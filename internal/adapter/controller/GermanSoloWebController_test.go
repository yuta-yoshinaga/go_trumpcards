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

func mustGermanSoloOutputJSON(msg string) string {
	// **本番の既定値と同じ形にする。** 空スライスを nil で、-1 の席を 0 で
	// 書くと、エラー応答だけ「席 0 が味方」「呼べるエースが未定義」に見える。
	out := &controller.GermanSoloWebOutput{
		Players:          []*controller.GermanSoloWebOutputPlayer{},
		CurrentTrick:     []*controller.WebOutputTrickCard{},
		PlayableIndices:  []int{},
		BiddableBids:     []int{},
		CallableAceSuits: []int{},
		CalledAceSuit:    -1,
		PartnerIdx:       -1,
		DeclarerIdx:      -1,
		TrumpSuit:        -1,
		LastTrickWinner:  -1,
		WinnerPlayer:     -1,
		WebOutputBase:    controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGermanSoloOutputJSON: %v", err))
	}
	return string(b)
}

func TestGermanSoloWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockGermanSoloInteractor)
	diMock.On("ResetWithConfig", domain.DefaultGermanSoloConfig()).Return(mockOutput)
	diMock.On("Bid", domain.GermanSoloBidFrage, domain.CardDesignHeart).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.GermanSoloInteractorIF { return diMock }
	ctrl := controller.NewGermanSoloWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.GermanSoloWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustGermanSoloOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid", func(t *testing.T) {
		input := controller.GermanSoloWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Bid:          func() *int { v := int(domain.GermanSoloBidFrage); return &v }(),
			TrumpSuit:    func() *int { v := domain.CardDesignHeart; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("bid missing bid", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`, mustGermanSoloOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.GermanSoloWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustGermanSoloOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
		run(t, `{"command":"other","sessionId":"s1"}`, mustGermanSoloOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustGermanSoloOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestGermanSoloWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		diff := 2
		rounds := 7
		expected := domain.GermanSoloConfig{CpuDifficulty: domain.GermanSoloCpuDifficultyHard, TargetRounds: 7}
		diMock := new(usecase.MockGermanSoloInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewGermanSoloWebController(func() uc.GermanSoloInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.GermanSoloWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.GermanSoloWebConfig{CpuDifficulty: &diff, TargetRounds: &rounds},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultGermanSoloConfig()
		diMock := new(usecase.MockGermanSoloInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewGermanSoloWebController(func() uc.GermanSoloInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.GermanSoloWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.GermanSoloWebConfig{CpuDifficulty: &diff},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultGermanSoloConfig()
		diMock := new(usecase.MockGermanSoloInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewGermanSoloWebController(func() uc.GermanSoloInteractorIF { return diMock })
		defer ctrl.Stop()

		input := controller.GermanSoloWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestGermanSoloWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockGermanSoloInteractor)
	c := controller.NewGermanSoloWebController(func() uc.GermanSoloInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

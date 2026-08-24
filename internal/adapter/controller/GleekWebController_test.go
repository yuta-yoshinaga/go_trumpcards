//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustGleekOutputJSON(msg string) string {
	// **本番の既定値と同じ形にする。** 空スライスを nil で、-1 の席を 0 で
	// 書くと、エラー応答だけ「席 0 が落札者」に見える。
	out := &controller.GleekWebOutput{
		Players:         []*controller.GleekWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		Melds:           []*controller.GleekWebOutputMeld{},
		BuyerIdx:        -1,
		RuffWinnerIdx:   -1,
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		DiscardCount:    domain.GleekSwapSize,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGleekOutputJSON: %v", err))
	}
	return string(b)
}

func TestGleekWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockGleekInteractor)
	diMock.On("ResetWithConfig", domain.DefaultGleekConfig()).Return(mockOutput)
	diMock.On("Bid", 14).Return(mockOutput)
	diMock.On("Bid", 0).Return(mockOutput)
	diMock.On("Discard", []int{0, 1, 2, 3, 4, 5, 6}).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)
	factory := func() uc.GleekInteractorIF { return diMock }
	ctrl := controller.NewGleekWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.GleekWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"r","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid raises", func(t *testing.T) {
		input := controller.GleekWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Bid:          func() *int { v := 14; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	// **0 は「降りる」で、省略ではない。** nil と 0 を同一視すると、降りる操作が
	// 「bid is required」で弾かれる。
	t.Run("bid zero drops out rather than erroring", func(t *testing.T) {
		run(t, `{"command":"bid","bid":0,"sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid missing bid", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`, mustGleekOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("discard", func(t *testing.T) {
		run(t, `{"command":"d","discardIndices":[0,1,2,3,4,5,6],"sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("discard missing indices", func(t *testing.T) {
		run(t, `{"command":"d","sessionId":"s1"}`, mustGleekOutputJSON("param error: discardIndices is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.GleekWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`, mustGleekOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("next / nextround / log / hint", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustGleekOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustGleekOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestGleekWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	t.Run("custom config passed through", func(t *testing.T) {
		expected := domain.DefaultGleekConfig()
		expected.CpuDifficulty = domain.GleekCpuDifficultyHard
		expected.TargetRounds = 7

		diMock := new(usecase.MockGleekInteractor)
		diMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		ctrl := controller.NewGleekWebController(func() uc.GleekInteractorIF { return diMock })
		defer ctrl.Stop()

		var input controller.GleekWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1","config":{"cpuDifficulty":2,"targetRounds":7}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	// **範囲外の設定は既定に丸める。** 素通しすると、CPU 難易度が定義の外に出る。
	t.Run("out-of-range config falls back to the default", func(t *testing.T) {
		diMock := new(usecase.MockGleekInteractor)
		diMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		ctrl := controller.NewGleekWebController(func() uc.GleekInteractorIF { return diMock })
		defer ctrl.Stop()

		var input controller.GleekWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1","config":{"cpuDifficulty":99,"targetRounds":0}}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", domain.DefaultGleekConfig())
	})
}

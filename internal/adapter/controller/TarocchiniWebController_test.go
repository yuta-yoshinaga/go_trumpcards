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

func mustTarocchiniOutputJSON(msg string) string {
	out := &controller.TarocchiniWebOutput{
		Players:         []*controller.TarocchiniWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTarocchiniOutputJSON: %v", err))
	}
	return string(b)
}

func TestTarocchiniWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockTarocchiniInteractor)
	diMock.On("ResetWithConfig", domain.DefaultTarocchiniConfig()).Return(mockOutput)
	diMock.On("Discard", []int{0, 1}).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewTarocchiniWebController(func() uc.TarocchiniInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.TarocchiniWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTarocchiniOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("scarto and its aliases", func(t *testing.T) {
		for _, cmd := range []string{"s", "scarto", "d", "discard"} {
			input := controller.TarocchiniWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "s1"},
				CardIndices:  []int{0, 1},
			}
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})
	t.Run("scarto missing cardIndices", func(t *testing.T) {
		run(t, `{"command":"s","sessionId":"s1"}`,
			mustTarocchiniOutputJSON("param error: cardIndices is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.TarocchiniWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`,
			mustTarocchiniOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	// 入札は存在しないので、bid は素通りせず弾かれる必要がある。
	t.Run("bid is not a command here", func(t *testing.T) {
		run(t, `{"command":"bid","sessionId":"s1"}`,
			mustTarocchiniOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("next / nextround / log / hint", func(t *testing.T) {
		for _, cmd := range []string{"n", "nr", "log", "h"} {
			run(t, fmt.Sprintf(`{"command":%q,"sessionId":"s1"}`, cmd), mockOutput, http.StatusOK)
		}
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`,
			mustTarocchiniOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`,
			mustTarocchiniOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestTarocchiniWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	newCtrl := func(expected domain.TarocchiniConfig) (*usecase.MockTarocchiniInteractor, *controller.TarocchiniWebController) {
		diMock := new(usecase.MockTarocchiniInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		return diMock, controller.NewTarocchiniWebController(func() uc.TarocchiniInteractorIF { return diMock })
	}

	t.Run("custom config passed through", func(t *testing.T) {
		diff, rounds := 2, 8
		expected := domain.TarocchiniConfig{CpuDifficulty: domain.TarocchiniCpuDifficultyHard, TargetRounds: 8}
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.TarocchiniWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.TarocchiniWebConfig{CpuDifficulty: &diff, TargetRounds: &rounds},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultTarocchiniConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.TarocchiniWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.TarocchiniWebConfig{CpuDifficulty: &diff},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	// プレイヤー数未満の局数は境界検査を通っても Validate が落とすので、
	// ここで既定に丸めておかないとリセットが黙って無視される。
	t.Run("rounds below the player count fall back to default", func(t *testing.T) {
		rounds := 1
		expected := domain.DefaultTarocchiniConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.TarocchiniWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
			Config:       &controller.TarocchiniWebConfig{TargetRounds: &rounds},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultTarocchiniConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.TarocchiniWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c4"},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestTarocchiniWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockTarocchiniInteractor)
	c := controller.NewTarocchiniWebController(func() uc.TarocchiniInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

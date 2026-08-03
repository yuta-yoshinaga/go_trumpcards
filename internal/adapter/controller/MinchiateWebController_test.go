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

// minchiateSurplusIndices は捨てる枚数ぶんの位置を返す。
//
// **枚数は定数から出す。**Minchiate の余剰は 13 枚で、Tarocchini の 2 枚では
// ない。テストに数字を直書きすると誤った枚数が仕様として読まれる。
func minchiateSurplusIndices() []int {
	idx := make([]int, 0, domain.MinchiateSurplus)
	for i := 0; i < domain.MinchiateSurplus; i++ {
		idx = append(idx, i)
	}
	return idx
}

func mustMinchiateOutputJSON(msg string) string {
	out := &controller.MinchiateWebOutput{
		Players:         []*controller.MinchiateWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMinchiateOutputJSON: %v", err))
	}
	return string(b)
}

func TestMinchiateWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockMinchiateInteractor)
	diMock.On("ResetWithConfig", domain.DefaultMinchiateConfig()).Return(mockOutput)
	diMock.On("Discard", minchiateSurplusIndices()).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewMinchiateWebController(func() uc.MinchiateInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.MinchiateWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustMinchiateOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("scarto and its aliases", func(t *testing.T) {
		for _, cmd := range []string{"s", "scarto", "d", "discard"} {
			input := controller.MinchiateWebInput{
				BaseWebInput: controller.BaseWebInput{Command: cmd, SessionID: "s1"},
				CardIndices:  minchiateSurplusIndices(),
			}
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})
	t.Run("scarto missing cardIndices", func(t *testing.T) {
		run(t, `{"command":"s","sessionId":"s1"}`,
			mustMinchiateOutputJSON("param error: cardIndices is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.MinchiateWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`,
			mustMinchiateOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	// 入札は存在しないので、bid は素通りせず弾かれる必要がある。
	t.Run("bid is not a command here", func(t *testing.T) {
		run(t, `{"command":"bid","sessionId":"s1"}`,
			mustMinchiateOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("next / nextround / log / hint", func(t *testing.T) {
		for _, cmd := range []string{"n", "nr", "log", "h"} {
			run(t, fmt.Sprintf(`{"command":%q,"sessionId":"s1"}`, cmd), mockOutput, http.StatusOK)
		}
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`,
			mustMinchiateOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty command", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`,
			mustMinchiateOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestMinchiateWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTrick":[]}`

	newCtrl := func(expected domain.MinchiateConfig) (*usecase.MockMinchiateInteractor, *controller.MinchiateWebController) {
		diMock := new(usecase.MockMinchiateInteractor)
		diMock.On("ResetWithConfig", expected).Return(mockOutput)
		return diMock, controller.NewMinchiateWebController(func() uc.MinchiateInteractorIF { return diMock })
	}

	t.Run("custom config passed through", func(t *testing.T) {
		diff, rounds := 2, 8
		expected := domain.MinchiateConfig{CpuDifficulty: domain.MinchiateCpuDifficultyHard, TargetRounds: 8}
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.MinchiateWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.MinchiateWebConfig{CpuDifficulty: &diff, TargetRounds: &rounds},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range difficulty falls back to default", func(t *testing.T) {
		diff := 9
		expected := domain.DefaultMinchiateConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.MinchiateWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.MinchiateWebConfig{CpuDifficulty: &diff},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	// プレイヤー数未満の局数は境界検査を通っても Validate が落とすので、
	// ここで既定に丸めておかないとリセットが黙って無視される。
	t.Run("rounds below the player count fall back to default", func(t *testing.T) {
		rounds := 1
		expected := domain.DefaultMinchiateConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.MinchiateWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
			Config:       &controller.MinchiateWebConfig{TargetRounds: &rounds},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultMinchiateConfig()
		diMock, ctrl := newCtrl(expected)
		defer ctrl.Stop()

		input := controller.MinchiateWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c4"},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestMinchiateWebController_Stop(t *testing.T) {
	diMock := new(usecase.MockMinchiateInteractor)
	c := controller.NewMinchiateWebController(func() uc.MinchiateInteractorIF { return diMock })
	c.Stop()
	c.Stop()
}

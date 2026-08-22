//go:build test

package controller_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func setupS27WebTest(t *testing.T) (*usecase.MockSevenTwentySevenInteractor, http.HandlerFunc, string) {
	t.Helper()
	mockOutput := `{"players":[],"phase":0,"message":""}`
	m := new(usecase.MockSevenTwentySevenInteractor)
	ctrl := controller.NewSevenTwentySevenWebController(func() uc.SevenTwentySevenInteractorIF { return m })
	t.Cleanup(func() { ctrl.Stop() })
	return m, ctrl.Exec, mockOutput
}

// **`card` は引く、`stand` は止まる。** どちらも追加パラメータ無しで通ること。
func TestSevenTwentySevenWebController_CardAndStand(t *testing.T) {
	for _, tt := range []struct {
		command string
		draw    bool
	}{{"card", true}, {"c", true}, {"stand", false}, {"s", false}} {
		t.Run(tt.command, func(t *testing.T) {
			m, handler, out := setupS27WebTest(t)
			m.On("TakeCard", tt.draw).Return(out)
			rec := execRequest(t, handler, &controller.SevenTwentySevenWebInput{
				BaseWebInput: controller.BaseWebInput{Command: tt.command, SessionID: "s1"},
			})
			assert.Equal(t, http.StatusOK, rec.Code)
			m.AssertCalled(t, "TakeCard", tt.draw)
		})
	}
}

// **Guts の declare コマンドは残っていない。** 残すと、送れば効きそうに見えて
// 何も起きない入力になる。
func TestSevenTwentySevenWebController_RejectsDeclare(t *testing.T) {
	m, handler, _ := setupS27WebTest(t)
	execRequest(t, handler, &controller.SevenTwentySevenWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "declare", SessionID: "s2"},
	})
	m.AssertNotCalled(t, "TakeCard", mock.Anything)
}

func TestSevenTwentySevenWebController_ResetAndNextRound(t *testing.T) {
	m, handler, out := setupS27WebTest(t)
	m.On("ResetWithConfig", mock.Anything).Return(out)
	m.On("NextRound").Return(out)

	execRequest(t, handler, &controller.SevenTwentySevenWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s3"},
	})
	m.AssertCalled(t, "ResetWithConfig", mock.Anything)

	execRequest(t, handler, &controller.SevenTwentySevenWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "nextround", SessionID: "s3"},
	})
	m.AssertCalled(t, "NextRound")
}

// **設定は reset のときだけ効く。** ToConfig が境界を丸めていること。
func TestSevenTwentySevenWebController_ToConfigBounds(t *testing.T) {
	tooSmall, tooBig := 0, 99
	in := controller.SevenTwentySevenWebInput{
		Config: &controller.SevenTwentySevenWebConfig{
			PlayerCount:   &tooBig,
			Ante:          &tooSmall,
			StartingChips: &tooSmall,
			TargetRounds:  &tooSmall,
		},
	}
	cfg := in.ToConfig()
	def := domain.DefaultSevenTwentySevenConfig()
	assert.LessOrEqual(t, cfg.PlayerCount, domain.SevenTwentySevenMaxPlayerCount)
	assert.GreaterOrEqual(t, cfg.PlayerCount, domain.SevenTwentySevenMinPlayerCount)
	assert.Positive(t, cfg.Ante)
	assert.Positive(t, cfg.StartingChips)
	assert.Positive(t, cfg.TargetRounds)

	// 設定が無ければ既定値。
	assert.Equal(t, def, controller.SevenTwentySevenWebInput{}.ToConfig())
}

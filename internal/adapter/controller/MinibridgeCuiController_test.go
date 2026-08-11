//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMinibridgeCui() (*controller.MinibridgeCuiController, *usecase.MockMinibridgeInteractor) {
	mi := new(usecase.MockMinibridgeInteractor)
	return controller.NewMinibridgeCuiController(mi), mi
}

func TestMinibridgeCuiController_Commands(t *testing.T) {
	c, mi := newMinibridgeCui()
	cfg := domain.DefaultMinibridgeConfig()
	mi.On("GetConfig").Return(cfg)
	mi.On("ResetWithConfig", cfg).Return("reset")
	mi.On("Contract", 3, 0).Return("contract")
	mi.On("Contract", 1, 4).Return("contract_suit")
	mi.On("Play", 2).Return("play")
	mi.On("NextRound").Return("next")
	mi.On("GiveUp").Return("giveup")
	mi.On("Hint").Return("hint")
	mi.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		// **ノートランプは 0。** 省略ではなく 5 つ目の選択肢。
		{"contract no-trump", "c 3 0", "contract"},
		{"contract a suit", "contract 1 4", "contract_suit"},
		{"play", "p 2", "play"},
		{"next", "n", "next"},
		{"giveup", "g", "giveup"},
		{"hint", "h", "hint"},
		{"log", "l", "log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
		})
	}
}

// **引数を既定値で埋めない。** 埋めると選んでいない契約を引き受けてしまう。
func TestMinibridgeCuiController_ContractRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"level missing", "c", "Level is required."},
		{"suit missing", "c 3", "Suit is required."},
		{"level not a number", "c x 0", "Invalid level: x."},
		{"suit not a number", "c 3 x", "Invalid suit: x."},
		{"level below the minimum", "c 0 0", "Invalid level: 0."},
		{"level above the maximum", "c 8 0", "Invalid level: 8."},
		{"suit below the minimum", "c 3 -1", "Invalid suit: -1."},
		{"suit above the maximum", "c 3 5", "Invalid suit: 5."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, mi := newMinibridgeCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			mi.AssertNotCalled(t, "Contract")
		})
	}
}

func TestMinibridgeCuiController_PlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"index missing", "p", "Card index is required."},
		{"index not a number", "p x", "Invalid card index: x."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, mi := newMinibridgeCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			mi.AssertNotCalled(t, "Play")
		})
	}
}

// **競りは無いので bid コマンドも無い。**
func TestMinibridgeCuiController_NoBidCommand(t *testing.T) {
	c, mi := newMinibridgeCui()
	assert.Contains(t, c.Exec("bid 3"), "bid")
	mi.AssertNotCalled(t, "Contract")
}

func TestMinibridgeCuiController_UnknownCommand(t *testing.T) {
	c, _ := newMinibridgeCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}

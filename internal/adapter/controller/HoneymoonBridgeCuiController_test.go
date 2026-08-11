//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newHoneymoonBridgeCui() (*controller.HoneymoonBridgeCuiController, *usecase.MockHoneymoonBridgeInteractor) {
	hi := new(usecase.MockHoneymoonBridgeInteractor)
	return controller.NewHoneymoonBridgeCuiController(hi), hi
}

func TestHoneymoonBridgeCuiController_Commands(t *testing.T) {
	c, hi := newHoneymoonBridgeCui()
	cfg := domain.DefaultHoneymoonBridgeConfig()
	hi.On("GetConfig").Return(cfg)
	hi.On("ResetWithConfig", cfg).Return("reset")
	hi.On("Bid", 3, 0).Return("bid")
	hi.On("Bid", 1, 4).Return("bid_suit")
	hi.On("Pass").Return("pass")
	hi.On("Play", 2).Return("play")
	hi.On("NextRound").Return("next")
	hi.On("GiveUp").Return("giveup")
	hi.On("Hint").Return("hint")
	hi.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		// **ノートランプは 0。** 省略ではなく 5 つ目の選択肢。
		{"bid no-trump", "b 3 0", "bid"},
		{"bid a suit", "bid 1 4", "bid_suit"},
		{"pass", "pass", "pass"},
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

// **引数を既定値で埋めない。** 埋めると宣言していない契約を落札してしまう。
func TestHoneymoonBridgeCuiController_BidRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"level missing", "b", "Level is required."},
		{"suit missing", "b 3", "Suit is required."},
		{"level not a number", "b x 0", "Invalid level: x."},
		{"suit not a number", "b 3 x", "Invalid suit: x."},
		{"level below the minimum", "b 0 0", "Invalid level: 0."},
		{"level above the maximum", "b 8 0", "Invalid level: 8."},
		{"suit below the minimum", "b 3 -1", "Invalid suit: -1."},
		{"suit above the maximum", "b 3 5", "Invalid suit: 5."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, hi := newHoneymoonBridgeCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			hi.AssertNotCalled(t, "Bid")
		})
	}
}

func TestHoneymoonBridgeCuiController_PlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"index missing", "p", "Card index is required."},
		{"index not a number", "p x", "Invalid card index: x."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, hi := newHoneymoonBridgeCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			hi.AssertNotCalled(t, "Play")
		})
	}
}

func TestHoneymoonBridgeCuiController_UnknownCommand(t *testing.T) {
	c, _ := newHoneymoonBridgeCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}

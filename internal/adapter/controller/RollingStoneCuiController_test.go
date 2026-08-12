//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newRollingStoneCui() (*controller.RollingStoneCuiController, *usecase.MockRollingStoneInteractor) {
	ri := new(usecase.MockRollingStoneInteractor)
	return controller.NewRollingStoneCuiController(ri), ri
}

func TestRollingStoneCuiController_Commands(t *testing.T) {
	c, ri := newRollingStoneCui()
	cfg := domain.DefaultRollingStoneConfig()
	ri.On("GetConfig").Return(cfg)
	ri.On("ResetWithConfig", cfg).Return("reset")
	ri.On("Play", 2).Return("play")
	ri.On("PickUp").Return("pickup")
	ri.On("GiveUp").Return("giveup")
	ri.On("Hint").Return("hint")
	ri.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		{"play", "p 2", "play"},
		// **引き取りは別のコマンドで、引数を取らない。**
		{"pickup", "u", "pickup"},
		{"pickup long", "pickup", "pickup"},
		{"giveup", "g", "giveup"},
		{"hint", "h", "hint"},
		{"log", "l", "log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
		})
	}
}

func TestRollingStoneCuiController_PlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"index missing", "p", "Card index is required."},
		{"index not a number", "p x", "Invalid card index: x."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, ri := newRollingStoneCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			ri.AssertNotCalled(t, "Play")
		})
	}
}

func TestRollingStoneCuiController_UnknownCommand(t *testing.T) {
	c, _ := newRollingStoneCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}

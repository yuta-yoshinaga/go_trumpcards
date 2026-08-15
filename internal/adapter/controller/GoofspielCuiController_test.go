//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newGoofspielCui() (*controller.GoofspielCuiController, *usecase.MockGoofspielInteractor) {
	gi := new(usecase.MockGoofspielInteractor)
	return controller.NewGoofspielCuiController(gi), gi
}

func TestGoofspielCuiController_Commands(t *testing.T) {
	c, gi := newGoofspielCui()
	cfg := domain.DefaultGoofspielConfig()
	gi.On("GetConfig").Return(cfg)
	gi.On("ResetWithConfig", cfg).Return("reset")
	gi.On("Bid", 2).Return("bid")
	gi.On("NextRound").Return("next")
	gi.On("GiveUp").Return("giveup")
	gi.On("Hint").Return("hint")
	gi.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		{"bid", "b 2", "bid"},
		{"bid long", "bid 2", "bid"},
		{"next", "n", "next"},
		{"next long", "next", "next"},
		{"giveup", "g", "giveup"},
		{"hint", "h", "hint"},
		{"log", "l", "log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
		})
	}
}

func TestGoofspielCuiController_BidRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"index missing", "b", msgCardIndexRequired()},
		{"index not a number", "b x", msgInvalidCardIndex("x")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, gi := newGoofspielCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			gi.AssertNotCalled(t, "Bid")
		})
	}
}

func TestGoofspielCuiController_UnknownCommand(t *testing.T) {
	c, _ := newGoofspielCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}

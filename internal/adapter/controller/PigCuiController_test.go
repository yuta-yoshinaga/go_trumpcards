//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newPigCui() (*controller.PigCuiController, *usecase.MockPigInteractor) {
	pi := new(usecase.MockPigInteractor)
	return controller.NewPigCuiController(pi), pi
}

func TestPigCuiController_Commands(t *testing.T) {
	c, pi := newPigCui()
	cfg := domain.DefaultPigConfig()
	pi.On("GetConfig").Return(cfg)
	pi.On("ResetWithConfig", cfg).Return("reset")
	pi.On("Pass", 2).Return("pass")
	pi.On("Signal").Return("signal")
	pi.On("NextRound").Return("next")
	pi.On("GiveUp").Return("giveup")
	pi.On("Hint").Return("hint")
	pi.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		{"pass", "p 2", "pass"},
		{"pass long", "pass 2", "pass"},
		// **合図は引数を取らない別のコマンド。**
		{"signal", "s", "signal"},
		{"signal long", "signal", "signal"},
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

func TestPigCuiController_PassRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"index missing", "p", msgCardIndexRequired()},
		{"index not a number", "p x", msgInvalidCardIndex("x")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, pi := newPigCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			pi.AssertNotCalled(t, "Pass")
		})
	}
}

// **合図は引数を無視します。** 気づいたかどうかに札は関係ありません。
func TestPigCuiController_SignalIgnoresArguments(t *testing.T) {
	c, pi := newPigCui()
	pi.On("Signal").Return("signal")
	assert.Equal(t, "signal", c.Exec("s 3"))
}

func TestPigCuiController_UnknownCommand(t *testing.T) {
	c, _ := newPigCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}

//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newPasurCui() (*controller.PasurCuiController, *usecase.MockPasurInteractor) {
	pi := new(usecase.MockPasurInteractor)
	return controller.NewPasurCuiController(pi), pi
}

func TestPasurCuiController_Commands(t *testing.T) {
	c, pi := newPasurCui()
	cfg := domain.DefaultPasurConfig()
	pi.On("GetConfig").Return(cfg)
	pi.On("ResetWithConfig", cfg).Return("reset")
	pi.On("Play", 2, []int{0, 3}).Return("capture")
	pi.On("Play", 2, []int{}).Return("trail")
	pi.On("GiveUp").Return("giveup")
	pi.On("Hint").Return("hint")
	pi.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		{"play and capture", "p 2 0 3", "capture"},
		// **場札の指定が無ければトレール。** 引数不足のエラーではない。
		{"play as a trail", "play 2", "trail"},
		{"giveup", "g", "giveup"},
		{"hint", "h", "hint"},
		{"log", "l", "log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
		})
	}
}

func TestPasurCuiController_PlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"index missing", "p", msgCardIndexRequired()},
		{"index not a number", "p x", msgInvalidCardIndex("x")},
		{"table index not a number", "p 1 x", msgKey("invalidTableIndexDot", "val", "x")},
		{"a later table index is bad", "p 1 0 y", msgKey("invalidTableIndexDot", "val", "y")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, pi := newPasurCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			pi.AssertNotCalled(t, "Play")
		})
	}
}

func TestPasurCuiController_UnknownCommand(t *testing.T) {
	c, _ := newPasurCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}

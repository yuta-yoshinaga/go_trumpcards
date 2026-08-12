//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSnapCuiController_Commands(t *testing.T) {
	si := new(usecase.MockSnapInteractor)
	c := controller.NewSnapCuiController(si)

	cfg := domain.DefaultSnapConfig()
	si.On("GetConfig").Return(cfg)
	si.On("ResetWithConfig", cfg).Return("reset")
	si.On("Step").Return("step")
	si.On("Snap").Return("snap")
	si.On("Tick").Return("tick")
	si.On("GiveUp").Return("giveup")
	si.On("Hint").Return("hint")
	si.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		{"step", "s", "step"},
		{"step long", "step", "step"},
		// **宣言は引数を取らない。** 席を選べると CPU に誤宣言させられる。
		{"snap", "n", "snap"},
		{"snap long", "snap", "snap"},
		{"tick", "t", "tick"},
		{"giveup", "g", "giveup"},
		{"hint", "h", "hint"},
		{"log", "l", "log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
		})
	}
}

// **余分な引数は無視する。** 席を指定する経路を作らない。
func TestSnapCuiController_SnapIgnoresArguments(t *testing.T) {
	si := new(usecase.MockSnapInteractor)
	c := controller.NewSnapCuiController(si)
	si.On("Snap").Return("snap")

	assert.Equal(t, "snap", c.Exec("n 1"))
	si.AssertNumberOfCalls(t, "Snap", 1)
}

func TestSnapCuiController_UnknownCommand(t *testing.T) {
	si := new(usecase.MockSnapInteractor)
	c := controller.NewSnapCuiController(si)
	assert.Contains(t, c.Exec("nope"), "nope")
}

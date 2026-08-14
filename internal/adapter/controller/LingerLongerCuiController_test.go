//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newLingerLongerCui() (*controller.LingerLongerCuiController, *usecase.MockLingerLongerInteractor) {
	li := new(usecase.MockLingerLongerInteractor)
	return controller.NewLingerLongerCuiController(li), li
}

func TestLingerLongerCuiController_Commands(t *testing.T) {
	c, li := newLingerLongerCui()
	cfg := domain.DefaultLingerLongerConfig()
	li.On("GetConfig").Return(cfg)
	li.On("ResetWithConfig", cfg).Return("reset")
	li.On("Play", 2).Return("play")
	li.On("GiveUp").Return("giveup")
	li.On("Hint").Return("hint")
	li.On("ActionLog").Return("log")

	for _, tc := range []struct{ name, cmd, want string }{
		{"quit", "q", "bye."},
		{"reset", "r", "reset"},
		{"play", "p 2", "play"},
		{"play long", "play 2", "play"},
		{"giveup", "g", "giveup"},
		{"hint", "h", "hint"},
		{"log", "l", "log"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
		})
	}
}

// **補充するコマンドは無い。** 取れば自動で 1 枚引くので、`draw` を受けてしまうと
// 「引かないと補充されない」という誤ったモデルを教えることになります。
func TestLingerLongerCuiController_HasNoDrawCommand(t *testing.T) {
	c, li := newLingerLongerCui()
	for _, cmd := range []string{"d", "draw", "u", "pickup"} {
		t.Run(cmd, func(t *testing.T) {
			assert.Contains(t, c.Exec(cmd), cmd)
		})
	}
	li.AssertNotCalled(t, "Play")
}

func TestLingerLongerCuiController_PlayRejectsBadArgs(t *testing.T) {
	for _, tc := range []struct{ name, cmd, want string }{
		{"index missing", "p", "Card index is required."},
		{"index not a number", "p x", "Invalid card index: x."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, li := newLingerLongerCui()
			assert.Equal(t, tc.want, c.Exec(tc.cmd))
			li.AssertNotCalled(t, "Play")
		})
	}
}

func TestLingerLongerCuiController_UnknownCommand(t *testing.T) {
	c, _ := newLingerLongerCui()
	assert.Contains(t, c.Exec("nope"), "nope")
}

//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	ucmock "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newS27CuiController() (*controller.SevenTwentySevenCuiController, *ucmock.MockSevenTwentySevenInteractor) {
	m := new(ucmock.MockSevenTwentySevenInteractor)
	return controller.NewSevenTwentySevenCuiController(m), m
}

// **`c` は引く、`s` は止まる。** Guts の `i`/`o` を残すと、打てるのに何も
// 起きないコマンドになる。
func TestSevenTwentySevenCuiController_CardAndStand(t *testing.T) {
	for _, tt := range []struct {
		cmd  string
		draw bool
	}{{"c", true}, {"card", true}, {"s", false}, {"stand", false}} {
		t.Run(tt.cmd, func(t *testing.T) {
			c, m := newS27CuiController()
			m.On("TakeCard", tt.draw).Return("ok")
			assert.Equal(t, "ok", c.Exec(tt.cmd))
			m.AssertCalled(t, "TakeCard", tt.draw)
		})
	}
}

// **入札系のコマンドは受け付けない。** Guts から来た i/o は存在しない。
func TestSevenTwentySevenCuiController_RejectsTheGutsCommands(t *testing.T) {
	for _, cmd := range []string{"i", "in", "o", "out"} {
		t.Run(cmd, func(t *testing.T) {
			c, m := newS27CuiController()
			out := c.Exec(cmd)
			assert.NotEmpty(t, out)
			m.AssertNotCalled(t, "TakeCard", mock.Anything)
		})
	}
}

func TestSevenTwentySevenCuiController_NextRound(t *testing.T) {
	c, m := newS27CuiController()
	m.On("NextRound").Return("next")
	assert.Equal(t, "next", c.Exec("n"))
	m.AssertCalled(t, "NextRound")
}

func TestSevenTwentySevenCuiController_HintAndLog(t *testing.T) {
	c, m := newS27CuiController()
	m.On("Hint").Return("hint")
	m.On("ActionLog").Return("log")
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("l"))
}

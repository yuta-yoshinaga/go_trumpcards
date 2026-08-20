//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newSpeedCuiMock() *usecase.MockSpeedInteractor {
	m := new(usecase.MockSpeedInteractor)
	m.On("ResetWithConfig", mock.Anything).Return("reset-ok")
	m.On("Play", mock.Anything, mock.Anything).Return("play-ok")
	m.On("Flip").Return("flip-ok")
	m.On("Hint").Return("hint-ok")
	m.On("ActionLog").Return("log-ok")
	m.On("GetConfig").Return(domain.DefaultSpeedConfig())
	return m
}

func TestSpeedCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		assert.Equal(t, i18n.QuitSentinel, c.Exec("q"))
		assert.Equal(t, i18n.QuitSentinel, c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		assert.Equal(t, "reset-ok", c.Exec("r"))
		assert.Equal(t, "reset-ok", c.Exec("reset"))
	})

	t.Run("play", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		result := c.Exec("p 0 1")
		assert.Equal(t, "play-ok", result)
		m.AssertCalled(t, "Play", 0, 1)
	})

	t.Run("play full word", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		result := c.Exec("play 2 0")
		assert.Equal(t, "play-ok", result)
		m.AssertCalled(t, "Play", 2, 0)
	})

	t.Run("flip", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		assert.Equal(t, "flip-ok", c.Exec("f"))
		assert.Equal(t, "flip-ok", c.Exec("flip"))
	})

	t.Run("hint", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		assert.Equal(t, "hint-ok", c.Exec("h"))
		assert.Equal(t, "hint-ok", c.Exec("hint"))
	})

	t.Run("log", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		assert.Equal(t, "log-ok", c.Exec("log"))
		assert.Equal(t, "log-ok", c.Exec("l"))
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, "reset-ok", result)
	})

	t.Run("empty input", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		result := c.Exec("")
		assert.NotEmpty(t, result)
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newSpeedCuiMock()
		c := controller.NewSpeedCuiController(m)
		result := c.Exec("xyz")
		assert.NotEmpty(t, result)
	})
}

// #5390: `p abc` が 0 番を出し、`sd abc` が Normal でゲームを配り直していた。
func TestSpeedCuiController_RefusesMistypedOptionalArgs(t *testing.T) {
	t.Run("play refuses a mistyped card index", func(t *testing.T) {
		m := newSpeedCuiMock()
		out := controller.NewSpeedCuiController(m).Exec("p abc")
		assert.Equal(t, msgKey("invalidCardIndex", "val", "abc"), out)
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
	})

	t.Run("play refuses a mistyped pile index", func(t *testing.T) {
		m := newSpeedCuiMock()
		out := controller.NewSpeedCuiController(m).Exec("p 0 xyz")
		assert.Equal(t, msgKey("invalidPileIndex", "val", "xyz"), out)
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
	})

	t.Run("play with no argument still plays the first card onto the first pile", func(t *testing.T) {
		m := newSpeedCuiMock()
		assert.Equal(t, "play-ok", controller.NewSpeedCuiController(m).Exec("p"))
		m.AssertCalled(t, "Play", 0, 0)
	})

	// **設定コマンドは特に静か。** 断らないと打ち間違いが Normal での配り直しに
	// なり、盤面が変わったこと以外に手掛かりが残らない。
	t.Run("setdifficulty refuses a mistyped level", func(t *testing.T) {
		m := newSpeedCuiMock()
		out := controller.NewSpeedCuiController(m).Exec("sd abc")
		assert.Equal(t, msgKey("invalidCpuDifficulty", "val", "abc"), out)
		m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
	})
}

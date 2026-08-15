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

func newWarCuiMock() *usecase.MockWarInteractor {
	m := new(usecase.MockWarInteractor)
	m.On("ResetWithConfig", mock.Anything).Return("reset-ok")
	m.On("Step").Return("step-ok")
	m.On("AutoPlay").Return("autoplay-ok")
	m.On("ActionLog").Return("log-ok")
	m.On("GetConfig").Return(domain.DefaultWarConfig())
	return m
}

func TestWarCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		m := newWarCuiMock()
		c := controller.NewWarCuiController(m)
		assert.Equal(t, i18n.QuitSentinel, c.Exec("q"))
		assert.Equal(t, i18n.QuitSentinel, c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newWarCuiMock()
		c := controller.NewWarCuiController(m)
		assert.Equal(t, "reset-ok", c.Exec("r"))
		assert.Equal(t, "reset-ok", c.Exec("reset"))
	})

	t.Run("step", func(t *testing.T) {
		m := newWarCuiMock()
		c := controller.NewWarCuiController(m)
		assert.Equal(t, "step-ok", c.Exec("s"))
		assert.Equal(t, "step-ok", c.Exec("step"))
	})

	t.Run("autoplay", func(t *testing.T) {
		m := newWarCuiMock()
		c := controller.NewWarCuiController(m)
		assert.Equal(t, "autoplay-ok", c.Exec("a"))
		assert.Equal(t, "autoplay-ok", c.Exec("autoplay"))
	})

	t.Run("setmax", func(t *testing.T) {
		m := newWarCuiMock()
		c := controller.NewWarCuiController(m)
		assert.Equal(t, "reset-ok", c.Exec("sm 1000"))
		assert.Equal(t, "reset-ok", c.Exec("setmax 1000"))
	})

	t.Run("log", func(t *testing.T) {
		m := newWarCuiMock()
		c := controller.NewWarCuiController(m)
		assert.Equal(t, "log-ok", c.Exec("log"))
		assert.Equal(t, "log-ok", c.Exec("l"))
	})

	t.Run("empty and unknown", func(t *testing.T) {
		m := newWarCuiMock()
		c := controller.NewWarCuiController(m)
		assert.NotEmpty(t, c.Exec(""))
		assert.NotEmpty(t, c.Exec("xyz"))
	})
}

// 打ち間違えた任意引数は既定値に落とさず断る (#5390)。`sm abc` が既定の
// ラウンド数で設定を上書きし直すと、局が黙って作り直される。
func TestWarCuiController_SetMaxRefusesMistypedValue(t *testing.T) {
	m := newWarCuiMock()
	c := controller.NewWarCuiController(m)
	assert.Equal(t, msgKey("invalidMaxRounds", "val", "abc"), c.Exec("sm abc"))
	m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)

	// 引数なしは既定値のまま。断る側だけ直して省略形を壊しても緑にならないように。
	assert.Equal(t, "reset-ok", c.Exec("sm"))
	m.AssertCalled(t, "ResetWithConfig", mock.Anything)
}

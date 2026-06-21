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

func newKempsCuiMock() *usecase.MockKempsInteractor {
	m := new(usecase.MockKempsInteractor)
	m.On("Reset").Return("reset-ok")
	m.On("Swap", mock.Anything, mock.Anything).Return("swap-ok")
	m.On("Pass").Return("pass-ok")
	m.On("SetSignal", mock.Anything).Return("signal-ok")
	m.On("DeclareKemps").Return("kemps-ok")
	m.On("DeclareCounterKemps", mock.Anything).Return("counter-ok")
	m.On("NextRound").Return("next-ok")
	m.On("ActionLog").Return("log-ok")
	m.On("GetConfig").Return(domain.DefaultKempsConfig())
	return m
}

func TestKempsCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		c := controller.NewKempsCuiController(newKempsCuiMock())
		assert.Equal(t, i18n.QuitSentinel, c.Exec("q"))
	})
	t.Run("reset", func(t *testing.T) {
		c := controller.NewKempsCuiController(newKempsCuiMock())
		assert.Equal(t, "reset-ok", c.Exec("r"))
		assert.Equal(t, "reset-ok", c.Exec("reset"))
	})
	t.Run("swap", func(t *testing.T) {
		m := newKempsCuiMock()
		c := controller.NewKempsCuiController(m)
		assert.Equal(t, "swap-ok", c.Exec("s 1 2"))
		assert.Equal(t, "swap-ok", c.Exec("swap 0 0"))
		m.AssertCalled(t, "Swap", 1, 2)
	})
	t.Run("pass", func(t *testing.T) {
		c := controller.NewKempsCuiController(newKempsCuiMock())
		assert.Equal(t, "pass-ok", c.Exec("p"))
		assert.Equal(t, "pass-ok", c.Exec("pass"))
	})
	t.Run("signal", func(t *testing.T) {
		m := newKempsCuiMock()
		c := controller.NewKempsCuiController(m)
		assert.Equal(t, "signal-ok", c.Exec("sig 1"))
		assert.Equal(t, "signal-ok", c.Exec("signal 0"))
		m.AssertCalled(t, "SetSignal", 1)
	})
	t.Run("kemps", func(t *testing.T) {
		c := controller.NewKempsCuiController(newKempsCuiMock())
		assert.Equal(t, "kemps-ok", c.Exec("k"))
		assert.Equal(t, "kemps-ok", c.Exec("kemps"))
	})
	t.Run("counter", func(t *testing.T) {
		m := newKempsCuiMock()
		c := controller.NewKempsCuiController(m)
		assert.Equal(t, "counter-ok", c.Exec("c 1"))
		assert.Equal(t, "counter-ok", c.Exec("counter 3"))
		m.AssertCalled(t, "DeclareCounterKemps", 1)
	})
	t.Run("next", func(t *testing.T) {
		c := controller.NewKempsCuiController(newKempsCuiMock())
		assert.Equal(t, "next-ok", c.Exec("n"))
		assert.Equal(t, "next-ok", c.Exec("next"))
	})
	t.Run("log", func(t *testing.T) {
		c := controller.NewKempsCuiController(newKempsCuiMock())
		assert.Equal(t, "log-ok", c.Exec("log"))
		assert.Equal(t, "log-ok", c.Exec("l"))
	})
	t.Run("empty and unknown", func(t *testing.T) {
		c := controller.NewKempsCuiController(newKempsCuiMock())
		assert.NotEmpty(t, c.Exec(""))
		assert.NotEmpty(t, c.Exec("xyz"))
	})
}

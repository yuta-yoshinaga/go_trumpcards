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

func newSlapjackCuiMock() *usecase.MockSlapjackInteractor {
	m := new(usecase.MockSlapjackInteractor)
	m.On("Reset").Return("reset-ok")
	m.On("Step").Return("step-ok")
	m.On("Slap", mock.Anything).Return("slap-ok")
	m.On("Tick").Return("tick-ok")
	m.On("ActionLog").Return("log-ok")
	m.On("GetConfig").Return(domain.DefaultSlapjackConfig())
	return m
}

func TestSlapjackCuiController_Exec(t *testing.T) {
	t.Run("quit", func(t *testing.T) {
		m := newSlapjackCuiMock()
		c := controller.NewSlapjackCuiController(m)
		assert.Equal(t, i18n.QuitSentinel, c.Exec("q"))
		assert.Equal(t, i18n.QuitSentinel, c.Exec("quit"))
	})

	t.Run("reset", func(t *testing.T) {
		m := newSlapjackCuiMock()
		c := controller.NewSlapjackCuiController(m)
		assert.Equal(t, "reset-ok", c.Exec("r"))
		assert.Equal(t, "reset-ok", c.Exec("reset"))
	})

	t.Run("step", func(t *testing.T) {
		m := newSlapjackCuiMock()
		c := controller.NewSlapjackCuiController(m)
		assert.Equal(t, "step-ok", c.Exec("s"))
		assert.Equal(t, "step-ok", c.Exec("step"))
	})

	t.Run("slap", func(t *testing.T) {
		m := newSlapjackCuiMock()
		c := controller.NewSlapjackCuiController(m)
		assert.Equal(t, "slap-ok", c.Exec("j"))
		assert.Equal(t, "slap-ok", c.Exec("slap"))
	})

	t.Run("tick", func(t *testing.T) {
		m := newSlapjackCuiMock()
		c := controller.NewSlapjackCuiController(m)
		assert.Equal(t, "tick-ok", c.Exec("tick"))
	})

	t.Run("log", func(t *testing.T) {
		m := newSlapjackCuiMock()
		c := controller.NewSlapjackCuiController(m)
		assert.Equal(t, "log-ok", c.Exec("log"))
		assert.Equal(t, "log-ok", c.Exec("l"))
	})

	t.Run("empty and unknown", func(t *testing.T) {
		m := newSlapjackCuiMock()
		c := controller.NewSlapjackCuiController(m)
		assert.NotEmpty(t, c.Exec(""))
		assert.NotEmpty(t, c.Exec("xyz"))
	})

	// `sd <n>` has been advertised by slapjack.helpSetDifficulty since the game
	// shipped but was never implemented -- it fell through to "Unknown command"
	// in both line and realtime mode, while the Web GUI has had a difficulty
	// selector all along. See issue #5179.
	t.Run("set difficulty", func(t *testing.T) {
		for _, tc := range []struct {
			cmd  string
			want domain.SlapjackCpuDifficulty
		}{
			{"sd 0", domain.SlapjackCpuEasy},
			{"sd 1", domain.SlapjackCpuNormal},
			{"sd 2", domain.SlapjackCpuHard},
			{"setdifficulty 2", domain.SlapjackCpuHard},
		} {
			m := newSlapjackCuiMock()
			var got domain.SlapjackConfig
			m.On("ResetWithConfig", mock.Anything).Return("reset-cfg-ok").Run(func(args mock.Arguments) {
				got = args.Get(0).(domain.SlapjackConfig)
			})
			c := controller.NewSlapjackCuiController(m)
			assert.Equal(t, "reset-cfg-ok", c.Exec(tc.cmd), tc.cmd)
			assert.Equal(t, tc.want, got.CpuDifficulty, tc.cmd)
		}
	})

	t.Run("set difficulty rejects bad input without resetting", func(t *testing.T) {
		for _, cmd := range []string{"sd", "sd x", "sd -1", "sd 3"} {
			m := newSlapjackCuiMock()
			m.On("ResetWithConfig", mock.Anything).Return("reset-cfg-ok")
			c := controller.NewSlapjackCuiController(m)
			out := c.Exec(cmd)
			assert.NotEmpty(t, out, cmd)
			assert.NotEqual(t, "reset-cfg-ok", out, cmd)
			m.AssertNotCalled(t, "ResetWithConfig", mock.Anything)
		}
	})
}

//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestMachiavelliCuiController_Exec(t *testing.T) {
	const mockOutput = `{"phase":0}`

	newMock := func() *mockUsecases.MockMachiavelliInteractor {
		m := new(mockUsecases.MockMachiavelliInteractor)
		m.On("GetConfig").Return(domain.DefaultMachiavelliConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Draw").Return(mockOutput)
		m.On("NewMeld", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewMachiavelliCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultMachiavelliConfig())
	})

	t.Run("draw aliases", func(t *testing.T) {
		for _, cmd := range []string{"dr", "draw"} {
			m := newMock()
			c := controller.NewMachiavelliCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "Draw")
		}
	})

	t.Run("newmeld", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nm 0 1 2"))
		m.AssertCalled(t, "NewMeld", []int{0, 1, 2})
	})

	t.Run("newmeld usage when too few", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		out := c.Exec("nm 0 1")
		assert.Contains(t, out, msgUsage("usageNmIJKAtLeast3HandIndices"))
		m.AssertNotCalled(t, "NewMeld", mock.Anything)
	})

	t.Run("layoff", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("lo 0 2"))
		m.AssertCalled(t, "Layoff", 0, 2)
	})

	t.Run("layoff usage when missing args", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		out := c.Exec("lo 0")
		assert.Contains(t, out, msgUsage("usageLoMeldidxHandindex"))
		m.AssertNotCalled(t, "Layoff", mock.Anything, mock.Anything)
	})

	t.Run("nextround aliases", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			m := newMock()
			c := controller.NewMachiavelliCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "NextRound")
		}
	})

	t.Run("setplayers", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pc 5"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.MachiavelliConfig) bool {
			return cfg.PlayerCount == 5
		}))
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.MachiavelliConfig) bool {
			return cfg.CpuDifficulty == domain.MachiavelliCpuDifficultyHard
		}))
	})

	t.Run("setrounds", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sr 5"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.MachiavelliConfig) bool {
			return cfg.TargetRounds == 5
		}))
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newMock()
		c := controller.NewMachiavelliCuiController(m)
		out := c.Exec("blarghhh")
		assert.NotEmpty(t, out)
	})
}

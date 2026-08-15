package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPishtiCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockPishtiInteractor {
		m := new(mockUsecases.MockPishtiInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultPishtiConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewPishtiCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("next", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("play command", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 2"))
		m.AssertCalled(t, "Play", 2)
	})

	t.Run("play missing arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		out := c.Exec("p")
		assert.Contains(t, out, "Usage:")
	})

	t.Run("play invalid arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		out := c.Exec("p xyz")
		assert.Contains(t, out, "Invalid")
	})

	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.PishtiConfig) bool {
			return cfg.CpuDifficulty == domain.PishtiDifficultyHard
		}))
	})

	t.Run("set difficulty invalid", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		out := c.Exec("sd 9")
		assert.Contains(t, out, msgInvalidCpuDifficultyPrefix())
	})

	t.Run("set players", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sp 3"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.PishtiConfig) bool {
			return cfg.PlayerCnt == 3
		}))
	})

	t.Run("set players invalid", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		out := c.Exec("sp 9")
		assert.True(t, msgRejected(out))
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		assert.Equal(t, "log", c.Exec("log"))
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newMock()
		c := controller.NewPishtiCuiController(m)
		out := c.Exec("zzz")
		assert.NotEmpty(t, out)
	})
}

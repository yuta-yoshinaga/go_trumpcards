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

func TestIndianRummyCuiController_Exec(t *testing.T) {
	const mockOutput = `{"phase":0}`

	newMock := func() *mockUsecases.MockIndianRummyInteractor {
		m := new(mockUsecases.MockIndianRummyInteractor)
		m.On("GetConfig").Return(domain.DefaultIndianRummyConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Declare", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewIndianRummyCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultIndianRummyConfig())
	})

	t.Run("drawstock aliases", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			m := newMock()
			c := controller.NewIndianRummyCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromStock")
		}
	})

	t.Run("drawdiscard aliases", func(t *testing.T) {
		for _, cmd := range []string{"dd", "drawdiscard"} {
			m := newMock()
			c := controller.NewIndianRummyCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromDiscard")
		}
	})

	t.Run("discard", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard usage when missing arg", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		out := c.Exec("d")
		assert.NotEmpty(t, out)
		m.AssertNotCalled(t, "Discard", mock.Anything)
	})

	t.Run("declare", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("de 2"))
		m.AssertCalled(t, "Declare", 2)
	})

	t.Run("nextround aliases", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			m := newMock()
			c := controller.NewIndianRummyCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "NextRound")
		}
	})

	t.Run("setplayers", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pc 3"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.IndianRummyConfig) bool {
			return cfg.PlayerCount == 3
		}))
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.IndianRummyConfig) bool {
			return cfg.CpuDifficulty == domain.IndianRummyCpuDifficultyHard
		}))
	})

	t.Run("setrounds", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sr 5"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.IndianRummyConfig) bool {
			return cfg.TargetRounds == 5
		}))
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newMock()
		c := controller.NewIndianRummyCuiController(m)
		out := c.Exec("blarghhh")
		assert.NotEmpty(t, out)
	})
}

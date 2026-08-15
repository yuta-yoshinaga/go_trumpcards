package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestScopaCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockScopaInteractor {
		m := new(mockUsecases.MockScopaInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultScopaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewScopaCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("next round", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("play with capture", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0 1 2"))
		m.AssertCalled(t, "Play", 0, []int{1, 2})
	})

	t.Run("play lay (no table)", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0"))
		m.AssertCalled(t, "Play", 0, []int{})
	})

	t.Run("play missing hand", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Contains(t, c.Exec("p"), msgUsage("usagePHandidxTableidxMany"))
	})

	t.Run("play bad hand index", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.True(t, msgRejected(c.Exec("p xyz")))
	})

	t.Run("sd (difficulty)", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.ScopaConfig) bool {
			return cfg.CpuDifficulty == domain.ScopaDifficultyHard
		}))
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewScopaCuiController(m)
		assert.Equal(t, "log", c.Exec("log"))
	})
}

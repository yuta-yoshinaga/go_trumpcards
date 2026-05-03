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

func TestPitchCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockPitchInteractor {
		m := new(mockUsecases.MockPitchInteractor)
		m.On("GetConfig").Return(domain.DefaultPitchConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewPitchCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})
	t.Run("quit quit", func(t *testing.T) {
		c := controller.NewPitchCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})
	t.Run("reset r", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultPitchConfig())
	})
	t.Run("bid pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("b 0"))
		m.AssertCalled(t, "Bid", 0)
	})
	t.Run("bid 3", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid 3"))
		m.AssertCalled(t, "Bid", 3)
	})
	t.Run("bid no args", func(t *testing.T) {
		c := controller.NewPitchCuiController(newMock())
		assert.NotEqual(t, mockOutput, c.Exec("b"))
	})
	t.Run("bid invalid value", func(t *testing.T) {
		c := controller.NewPitchCuiController(newMock())
		assert.NotEqual(t, mockOutput, c.Exec("b 99"))
	})
	t.Run("play 0", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0"))
		m.AssertCalled(t, "Play", 0)
	})
	t.Run("play no args", func(t *testing.T) {
		c := controller.NewPitchCuiController(newMock())
		assert.NotEqual(t, mockOutput, c.Exec("p"))
	})
	t.Run("next n", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})
	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("setdifficulty sd", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		expected := domain.DefaultPitchConfig()
		expected.CpuDifficulty = domain.PitchCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("setlimit sl", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 11"))
		expected := domain.DefaultPitchConfig()
		expected.PointLimit = 11
		m.AssertCalled(t, "ResetWithConfig", expected)
	})
	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("log l", func(t *testing.T) {
		m := newMock()
		c := controller.NewPitchCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewPitchCuiController(newMock())
		out := c.Exec("xyz")
		assert.NotEqual(t, mockOutput, out)
	})
}

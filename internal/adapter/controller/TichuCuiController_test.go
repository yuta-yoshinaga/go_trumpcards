package controller_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTichuCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockTichuInteractor {
		m := new(mockUsecases.MockTichuInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Declare", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultTichuConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return(`{"entries":[]}`)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewTichuCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewTichuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("play pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewTichuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p"))
		m.AssertCalled(t, "Play", []int{})
	})
	t.Run("play indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewTichuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 0 1"))
		m.AssertCalled(t, "Play", []int{0, 1})
	})
	t.Run("declare default", func(t *testing.T) {
		m := newMock()
		c := controller.NewTichuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d"))
		m.AssertCalled(t, "Declare", 0)
	})
	t.Run("declare value", func(t *testing.T) {
		m := newMock()
		c := controller.NewTichuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("declare 2"))
		m.AssertCalled(t, "Declare", 2)
	})
	t.Run("declare invalid", func(t *testing.T) {
		c := controller.NewTichuCuiController(newMock())
		assert.Contains(t, c.Exec("d 9"), "Invalid declaration")
	})
	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewTichuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 1"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("setdifficulty invalid", func(t *testing.T) {
		c := controller.NewTichuCuiController(newMock())
		assert.Contains(t, c.Exec("sd 9"), msgInvalidCpuDifficultyPrefix())
	})
	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewTichuCuiController(m)
		assert.Equal(t, `{"entries":[]}`, c.Exec("log"))
	})
}

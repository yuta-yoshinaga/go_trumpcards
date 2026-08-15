package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestShitheadCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockShitheadInteractor {
		m := new(mockUsecases.MockShitheadInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultShitheadConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewShitheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("play variants", func(t *testing.T) {
		m := newMock()
		c := controller.NewShitheadCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p"))
		m.AssertCalled(t, "Play", []int{})

		assert.Equal(t, mockOutput, c.Exec("p 2"))
		m.AssertCalled(t, "Play", []int{2})

		assert.Equal(t, mockOutput, c.Exec("play 0 1"))
		m.AssertCalled(t, "Play", []int{0, 1})
	})

	t.Run("set difficulty", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("set difficulty requires arg", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), msgCpuDifficultyRequired())
	})

	t.Run("set difficulty invalid arg", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		assert.Contains(t, c.Exec("sd 9"), msgInvalidCpuDifficultyPrefix())
	})

	t.Run("setrule list", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		result := c.Exec("sr list")
		assert.Contains(t, result, "two")
		assert.Contains(t, result, "seven")
		assert.Contains(t, result, "ten")
		assert.Contains(t, result, "fourburn")
	})

	t.Run("setrule usage", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		assert.Contains(t, c.Exec("sr"), "Usage")
	})

	t.Run("setrule unknown", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		assert.Contains(t, c.Exec("sr unknown 1"), "Unknown")
	})

	t.Run("setrule invalid value", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		assert.Contains(t, c.Exec("sr two 9"), "Invalid")
	})

	t.Run("setrule valid", func(t *testing.T) {
		c := controller.NewShitheadCuiController(newMock())
		result := c.Exec("sr two 0")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		m.On("ActionLog").Return("logs")
		c := controller.NewShitheadCuiController(m)
		assert.Equal(t, "logs", c.Exec("log"))
	})
}

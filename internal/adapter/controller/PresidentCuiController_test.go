package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPresidentCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockPresidentInteractor {
		m := new(mockUsecases.MockPresidentInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultPresidentConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewPresidentCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewPresidentCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint command", func(t *testing.T) {
		m := newMock()
		c := controller.NewPresidentCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("play command variants", func(t *testing.T) {
		m := newMock()
		c := controller.NewPresidentCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p"))
		m.AssertCalled(t, "Play", []int{})

		assert.Equal(t, mockOutput, c.Exec("p 2"))
		m.AssertCalled(t, "Play", []int{2})

		assert.Equal(t, mockOutput, c.Exec("play 0 1"))
		m.AssertCalled(t, "Play", []int{0, 1})
	})

	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewPresidentCuiController(m)
		result := c.Exec("sd 2")
		assert.Equal(t, mockOutput, result)
	})

	t.Run("set difficulty requires arg", func(t *testing.T) {
		c := controller.NewPresidentCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), "required")
	})

	t.Run("set difficulty invalid arg", func(t *testing.T) {
		c := controller.NewPresidentCuiController(newMock())
		assert.Contains(t, c.Exec("sd 9"), "Invalid")
	})

	t.Run("setrule list", func(t *testing.T) {
		c := controller.NewPresidentCuiController(newMock())
		result := c.Exec("sr list")
		assert.Contains(t, result, "revolution")
		assert.Contains(t, result, "exchange")
		assert.Contains(t, result, "passflush")
	})

	t.Run("setrule usage", func(t *testing.T) {
		c := controller.NewPresidentCuiController(newMock())
		assert.Contains(t, c.Exec("sr"), "Usage")
	})

	t.Run("setrule unknown", func(t *testing.T) {
		c := controller.NewPresidentCuiController(newMock())
		assert.Contains(t, c.Exec("sr unknown 1"), "Unknown")
	})

	t.Run("setrule invalid value", func(t *testing.T) {
		c := controller.NewPresidentCuiController(newMock())
		assert.Contains(t, c.Exec("sr revolution 9"), "Invalid")
	})

	t.Run("setrule valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewPresidentCuiController(m)
		result := c.Exec("sr revolution 0")
		assert.Equal(t, mockOutput, result)
	})
}

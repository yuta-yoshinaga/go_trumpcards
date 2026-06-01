package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBarbuCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockBarbuInteractor {
		m := new(mockUsecases.MockBarbuInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("NextDeal").Return(mockOutput)
		m.On("SelectContract", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultBarbuConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewBarbuCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBarbuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("next deal", func(t *testing.T) {
		m := newMock()
		c := controller.NewBarbuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextDeal")
	})

	t.Run("select contract with trump", func(t *testing.T) {
		m := newMock()
		c := controller.NewBarbuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 5 3"))
		m.AssertCalled(t, "SelectContract", 5, 3)
	})

	t.Run("select contract no trump", func(t *testing.T) {
		m := newMock()
		c := controller.NewBarbuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 0"))
		m.AssertCalled(t, "SelectContract", 0, -1)
	})

	t.Run("select contract missing arg", func(t *testing.T) {
		c := controller.NewBarbuCuiController(newMock())
		assert.Contains(t, c.Exec("c"), "Usage")
	})

	t.Run("play a card", func(t *testing.T) {
		m := newMock()
		c := controller.NewBarbuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 2"))
		m.AssertCalled(t, "Play", 2, []int(nil))
	})

	t.Run("play pass in dominoes", func(t *testing.T) {
		m := newMock()
		c := controller.NewBarbuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p -1"))
		m.AssertCalled(t, "Play", -1, []int(nil))
	})

	t.Run("play missing arg", func(t *testing.T) {
		c := controller.NewBarbuCuiController(newMock())
		assert.Contains(t, c.Exec("p"), "Usage")
	})

	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewBarbuCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("action log", func(t *testing.T) {
		m := newMock()
		c := controller.NewBarbuCuiController(m)
		assert.Equal(t, "log", c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
}

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
		assert.True(t, msgRejected(c.Exec("c")))
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
		assert.True(t, msgRejected(c.Exec("p")))
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

// codecov flagged both of Barbu's argument rejections.
func TestBarbuCuiController_ArgumentRejections(t *testing.T) {
	mockOutput := `{"players":[]}`
	newMock := func() *mockUsecases.MockBarbuInteractor {
		m := new(mockUsecases.MockBarbuInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("SelectContract", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		return m
	}

	t.Run("contract out of range", func(t *testing.T) {
		m := newMock()
		out := controller.NewBarbuCuiController(m).Exec("c 99")
		assert.Equal(t, msgKey("invalidContractRaw", "val", "99"), out)
		m.AssertNotCalled(t, "SelectContract", mock.Anything, mock.Anything)
	})

	t.Run("hand index is not a number", func(t *testing.T) {
		m := newMock()
		out := controller.NewBarbuCuiController(m).Exec("p abc")
		assert.Equal(t, msgKey("invalidHandIndexRaw", "val", "abc"), out)
		m.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
	})
}

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func TestDoubtCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockDoubtInteractor {
		m := new(mockUsecases.MockDoubtInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("ResolveDoubt", mock.Anything).Return(mockOutput)
		m.On("SkipDoubt").Return(mockOutput)
		return m
	}

	t.Run("quit command q", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit command quit", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset command r", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("r")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Reset")
	})

	t.Run("reset command reset", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("reset")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Reset")
	})

	t.Run("play command p with value and indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("p 5 0 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{0, 2}, 5)
	})

	t.Run("play command play with value and indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("play 3 0 1 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{0, 1, 2}, 3)
	})

	t.Run("play command p with no args uses 0 value and empty indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("p")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{}, 0)
	})

	t.Run("play command p with only value and no card indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("p 7")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "Play", []int{}, 7)
	})

	t.Run("doubt command d with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("d 1 2")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResolveDoubt", []int{1, 2})
	})

	t.Run("doubt command doubt with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("doubt 0")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResolveDoubt", []int{0})
	})

	t.Run("doubt command d with no indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("d")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "ResolveDoubt", []int{})
	})

	t.Run("skip command s", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("s")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "SkipDoubt")
	})

	t.Run("skip command skip", func(t *testing.T) {
		m := newMock()
		c := controller.NewDoubtCuiController(m)
		result := c.Exec("skip")
		assert.Equal(t, mockOutput, result)
		m.AssertCalled(t, "SkipDoubt")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("unknown")
		assert.Contains(t, result, "コマンドが不明です")
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewDoubtCuiController(newMock())
		result := c.Exec("")
		assert.Contains(t, result, "コマンドが不明です")
	})
}

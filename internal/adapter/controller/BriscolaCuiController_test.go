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

func TestBriscolaCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`
	newMock := func() *mockUsecases.MockBriscolaInteractor {
		m := new(mockUsecases.MockBriscolaInteractor)
		m.On("GetConfig").Return(domain.DefaultBriscolaConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit short", func(t *testing.T) {
		c := controller.NewBriscolaCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("quit long", func(t *testing.T) {
		c := controller.NewBriscolaCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBriscolaCuiController(m)
		got := c.Exec("r")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBriscolaConfig())
	})

	t.Run("play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBriscolaCuiController(m)
		got := c.Exec("p 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Play", 1)
	})

	t.Run("play missing index", func(t *testing.T) {
		c := controller.NewBriscolaCuiController(newMock())
		got := c.Exec("p")
		assert.Contains(t, got, msgCardIndexRequired())
	})

	t.Run("play invalid index", func(t *testing.T) {
		c := controller.NewBriscolaCuiController(newMock())
		got := c.Exec("p abc")
		assert.Contains(t, got, msgInvalidCardIndexPrefix())
	})

	t.Run("next short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBriscolaCuiController(m)
		got := c.Exec("n")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("hint short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBriscolaCuiController(m)
		got := c.Exec("h")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Hint")
	})

	t.Run("log short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBriscolaCuiController(m)
		got := c.Exec("l")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewBriscolaCuiController(newMock())
		got := c.Exec("xyz")
		assert.NotEqual(t, "bye.", got)
		assert.NotEmpty(t, got)
	})
}

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

func TestKingCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockKingInteractor {
		m := new(mockUsecases.MockKingInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("NextDeal").Return(mockOutput)
		m.On("SelectContract", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultKingConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit command", func(t *testing.T) {
		c := controller.NewKingCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewKingCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("next deal", func(t *testing.T) {
		m := newMock()
		c := controller.NewKingCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextDeal")
	})

	t.Run("select contract with trump", func(t *testing.T) {
		m := newMock()
		c := controller.NewKingCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 6 3"))
		m.AssertCalled(t, "SelectContract", 6, 3)
	})

	t.Run("select contract no trump", func(t *testing.T) {
		m := newMock()
		c := controller.NewKingCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("c 0"))
		m.AssertCalled(t, "SelectContract", 0, -1)
	})

	t.Run("select contract missing arg", func(t *testing.T) {
		c := controller.NewKingCuiController(newMock())
		assert.Contains(t, c.Exec("c"), "Usage")
	})

	t.Run("select contract invalid arg", func(t *testing.T) {
		c := controller.NewKingCuiController(newMock())
		assert.Contains(t, c.Exec("c abc"), msgStem("invalidContractRaw"))
	})

	t.Run("play a card", func(t *testing.T) {
		m := newMock()
		c := controller.NewKingCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("p 2"))
		m.AssertCalled(t, "Play", 2)
	})

	t.Run("play missing arg", func(t *testing.T) {
		c := controller.NewKingCuiController(newMock())
		assert.Contains(t, c.Exec("p"), "Usage")
	})

	t.Run("play invalid arg", func(t *testing.T) {
		c := controller.NewKingCuiController(newMock())
		assert.Contains(t, c.Exec("p xyz"), msgStem("invalidHandIndexRaw"))
	})

	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewKingCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewKingCuiController(m)
		assert.Equal(t, "hint", c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("action log", func(t *testing.T) {
		m := newMock()
		c := controller.NewKingCuiController(m)
		assert.Equal(t, "log", c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewKingCuiController(newMock())
		out := c.Exec("zzz")
		assert.NotEmpty(t, out)
	})
}

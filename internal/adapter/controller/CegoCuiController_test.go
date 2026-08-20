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

func TestCegoCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockCegoInteractor {
		m := new(mockUsecases.MockCegoInteractor)
		m.On("GetConfig").Return(domain.DefaultCegoConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("Pass").Return(mockOutput)
		m.On("ChooseContract", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewCegoCuiController(newMock()).Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultCegoConfig())
	})

	t.Run("bid play", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("bid play"))
		m.AssertCalled(t, "Bid", domain.CegoBidPlay)
	})

	t.Run("bid no args", func(t *testing.T) {
		result := controller.NewCegoCuiController(newMock()).Exec("bid")
		assert.Contains(t, result, msgStem("bidRequiredPlay"))
	})

	t.Run("bid invalid", func(t *testing.T) {
		result := controller.NewCegoCuiController(newMock()).Exec("bid zzz")
		assert.Contains(t, result, msgStem("invalidBidPlay"))
	})

	t.Run("pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pass"))
		m.AssertCalled(t, "Pass")
	})

	t.Run("cego contract", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("cego"))
		m.AssertCalled(t, "ChooseContract", domain.CegoContractCego)
	})

	t.Run("handspiel contract", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("handspiel"))
		m.AssertCalled(t, "ChooseContract", domain.CegoContractHandspiel)
	})

	t.Run("discard keep 1", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("discard 2"))
		m.AssertCalled(t, "Discard", []int{2})
	})

	t.Run("play", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("play 5"))
		m.AssertCalled(t, "Play", 5)
	})

	t.Run("next", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("n"))
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewCegoCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
}

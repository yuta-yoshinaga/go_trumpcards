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

func TestPrimeroCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockPrimeroInteractor {
		m := new(mockUsecases.MockPrimeroInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bet", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultPrimeroConfig())
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewPrimeroCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("call", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("c"))
		m.AssertCalled(t, "Bet", "call")
	})
	t.Run("raise", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("ra"))
		m.AssertCalled(t, "Bet", "raise")
	})
	t.Run("fold", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("fold"))
		m.AssertCalled(t, "Bet", "fold")
	})
	t.Run("next round", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set players", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("sp 3"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set players missing arg", func(t *testing.T) {
		out := controller.NewPrimeroCuiController(newMock()).Exec("sp")
		assert.True(t, msgRejected(out))
	})
	t.Run("set players invalid", func(t *testing.T) {
		out := controller.NewPrimeroCuiController(newMock()).Exec("sp 99")
		assert.True(t, msgRejected(out))
	})
	t.Run("set ante", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("sa 20"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set chips", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("sc 500"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set rounds", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewPrimeroCuiController(m).Exec("st 20"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "hint", controller.NewPrimeroCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("action log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "log", controller.NewPrimeroCuiController(m).Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewPrimeroCuiController(newMock()).Exec("zzz"))
	})
}

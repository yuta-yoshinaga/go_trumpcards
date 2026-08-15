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

func TestLooCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockLooInteractor {
		m := new(mockUsecases.MockLooInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Decide", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultLooConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewLooCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewLooCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("decide 1", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewLooCuiController(m).Exec("d 1"))
		m.AssertCalled(t, "Decide", true)
	})
	t.Run("decide 0", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewLooCuiController(m).Exec("d 0"))
		m.AssertCalled(t, "Decide", false)
	})
	t.Run("decide missing arg", func(t *testing.T) {
		out := controller.NewLooCuiController(newMock()).Exec("d")
		assert.Contains(t, out, "required")
	})
	t.Run("decide invalid arg", func(t *testing.T) {
		out := controller.NewLooCuiController(newMock()).Exec("d xyz")
		assert.Contains(t, out, "Invalid decision")
	})
	t.Run("play word", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewLooCuiController(m).Exec("play"))
		m.AssertCalled(t, "Decide", true)
	})
	t.Run("pass word", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewLooCuiController(m).Exec("pass"))
		m.AssertCalled(t, "Decide", false)
	})
	t.Run("play card", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewLooCuiController(m).Exec("p 2"))
		m.AssertCalled(t, "Play", 2)
	})
	t.Run("play invalid arg", func(t *testing.T) {
		out := controller.NewLooCuiController(newMock()).Exec("p xyz")
		assert.Contains(t, out, msgInvalidCardIndexPrefix())
	})
	t.Run("next round", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewLooCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewLooCuiController(m).Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "hint", controller.NewLooCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("action log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "log", controller.NewLooCuiController(m).Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewLooCuiController(newMock()).Exec("zzz"))
	})
}

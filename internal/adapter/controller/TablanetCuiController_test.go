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

func TestTablanetCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockTablanetInteractor {
		m := new(mockUsecases.MockTablanetInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultTablanetConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewTablanetCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTablanetCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("play hand only (trail)", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTablanetCuiController(m).Exec("p 1"))
		m.AssertCalled(t, "Play", 1, []int{})
	})
	t.Run("play with capture targets", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTablanetCuiController(m).Exec("p 0 1 2"))
		m.AssertCalled(t, "Play", 0, []int{1, 2})
	})
	t.Run("play missing index", func(t *testing.T) {
		out := controller.NewTablanetCuiController(newMock()).Exec("p")
		assert.Contains(t, out, msgCardIndexRequiredCapture())
	})
	t.Run("play invalid index", func(t *testing.T) {
		out := controller.NewTablanetCuiController(newMock()).Exec("p xyz")
		assert.Contains(t, out, msgInvalidCardIndexPrefix())
	})
	t.Run("next round", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTablanetCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewTablanetCuiController(m).Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "hint", controller.NewTablanetCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("action log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "log", controller.NewTablanetCuiController(m).Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewTablanetCuiController(newMock()).Exec("zzz"))
	})
}

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

func TestKoiKoiCuiController_Exec(t *testing.T) {
	mockOutput := `{"players":[]}`

	newMock := func() *mockUsecases.MockKoiKoiInteractor {
		m := new(mockUsecases.MockKoiKoiInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Play", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Decide", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultKoiKoiConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewKoiKoiCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewKoiKoiCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("play hand only", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewKoiKoiCuiController(m).Exec("p 1"))
		m.AssertCalled(t, "Play", 1, -1)
	})
	t.Run("play with field choice", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewKoiKoiCuiController(m).Exec("p 0 3"))
		m.AssertCalled(t, "Play", 0, 3)
	})
	t.Run("play missing index", func(t *testing.T) {
		out := controller.NewKoiKoiCuiController(newMock()).Exec("p")
		assert.Contains(t, out, msgCardIndexRequiredField())
	})
	t.Run("play invalid index", func(t *testing.T) {
		out := controller.NewKoiKoiCuiController(newMock()).Exec("p xyz")
		assert.Contains(t, out, msgInvalidCardIndexPrefix())
	})
	t.Run("koikoi", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewKoiKoiCuiController(m).Exec("kk"))
		m.AssertCalled(t, "Decide", true)
	})
	t.Run("shobu", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewKoiKoiCuiController(m).Exec("sb"))
		m.AssertCalled(t, "Decide", false)
	})
	t.Run("next round", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewKoiKoiCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set difficulty", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewKoiKoiCuiController(m).Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "hint", controller.NewKoiKoiCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("action log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "log", controller.NewKoiKoiCuiController(m).Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewKoiKoiCuiController(newMock()).Exec("zzz"))
	})
}

//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestMichiganCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockMichiganInteractor {
		m := new(mockUsecases.MockMichiganInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bet", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("GetConfig").Return(domain.DefaultMichiganConfig())
		m.On("Hint").Return("hint")
		m.On("ActionLog").Return("log")
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewMichiganCuiController(newMock()).Exec("q"))
	})
	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMichiganCuiController(m).Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("bet", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMichiganCuiController(m).Exec("b 2 2 2 2"))
		m.AssertCalled(t, "Bet", []int{2, 2, 2, 2})
	})
	t.Run("bet too few args", func(t *testing.T) {
		m := newMock()
		out := controller.NewMichiganCuiController(m).Exec("b 2 2")
		assert.NotEmpty(t, out)
		m.AssertNotCalled(t, "Bet", mock.Anything)
	})
	t.Run("bet invalid arg", func(t *testing.T) {
		m := newMock()
		out := controller.NewMichiganCuiController(m).Exec("b 2 x 2 2")
		assert.NotEmpty(t, out)
		m.AssertNotCalled(t, "Bet", mock.Anything)
	})
	t.Run("play", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMichiganCuiController(m).Exec("p 3"))
		m.AssertCalled(t, "Play", 3)
	})
	t.Run("play missing arg", func(t *testing.T) {
		out := controller.NewMichiganCuiController(newMock()).Exec("p")
		assert.Contains(t, out, i18n.T("cardIndexRequiredExample"))
	})
	t.Run("next round", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMichiganCuiController(m).Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})
	t.Run("set players", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMichiganCuiController(m).Exec("sp 5"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set players invalid", func(t *testing.T) {
		out := controller.NewMichiganCuiController(newMock()).Exec("sp 99")
		assert.True(t, msgRejected(out))
	})
	t.Run("set ante", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMichiganCuiController(m).Exec("sa 12"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set chips", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMichiganCuiController(m).Exec("sc 500"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("set rounds", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, mockOutput, controller.NewMichiganCuiController(m).Exec("st 20"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})
	t.Run("hint", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "hint", controller.NewMichiganCuiController(m).Exec("h"))
		m.AssertCalled(t, "Hint")
	})
	t.Run("action log", func(t *testing.T) {
		m := newMock()
		assert.Equal(t, "log", controller.NewMichiganCuiController(m).Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})
	t.Run("unknown", func(t *testing.T) {
		assert.NotEmpty(t, controller.NewMichiganCuiController(newMock()).Exec("zzz"))
	})
}

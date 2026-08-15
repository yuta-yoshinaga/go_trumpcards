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

func TestTrucoCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`
	newMock := func() *mockUsecases.MockTrucoInteractor {
		m := new(mockUsecases.MockTrucoInteractor)
		m.On("GetConfig").Return(domain.DefaultTrucoConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("Truco").Return(mockOutput)
		m.On("Respond", mock.Anything).Return(mockOutput)
		m.On("Next").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		assert.Equal(t, "bye.", controller.NewTrucoCuiController(newMock()).Exec("q"))
		assert.Equal(t, "bye.", controller.NewTrucoCuiController(newMock()).Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("r")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultTrucoConfig())
	})

	t.Run("play with index", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("p 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Play", 1)
	})

	t.Run("play missing index", func(t *testing.T) {
		got := controller.NewTrucoCuiController(newMock()).Exec("p")
		assert.Contains(t, got, msgCardIndexRequired())
	})

	t.Run("play invalid index", func(t *testing.T) {
		got := controller.NewTrucoCuiController(newMock()).Exec("p abc")
		assert.Contains(t, got, msgInvalidCardIndexPrefix())
	})

	t.Run("truco", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("t")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Truco")
	})

	t.Run("accept", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("a")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Respond", true)
	})

	t.Run("decline", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("d")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Respond", false)
	})

	t.Run("next", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("n")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Next")
	})

	t.Run("hint", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("h")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		got := controller.NewTrucoCuiController(m).Exec("l")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		got := controller.NewTrucoCuiController(newMock()).Exec("xyz")
		assert.NotEqual(t, "bye.", got)
		assert.NotEmpty(t, got)
	})
}

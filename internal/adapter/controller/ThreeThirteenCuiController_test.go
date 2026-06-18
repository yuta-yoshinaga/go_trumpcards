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

func TestThreeThirteenCuiController_Exec(t *testing.T) {
	const mockOutput = `{"phase":0}`

	newMock := func() *mockUsecases.MockThreeThirteenInteractor {
		m := new(mockUsecases.MockThreeThirteenInteractor)
		m.On("GetConfig").Return(domain.DefaultThreeThirteenConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("Knock", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewThreeThirteenCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultThreeThirteenConfig())
	})

	t.Run("drawstock aliases", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			m := newMock()
			c := controller.NewThreeThirteenCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromStock")
		}
	})

	t.Run("drawdiscard aliases", func(t *testing.T) {
		for _, cmd := range []string{"dd", "drawdiscard"} {
			m := newMock()
			c := controller.NewThreeThirteenCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromDiscard")
		}
	})

	t.Run("discard", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 0"))
		m.AssertCalled(t, "Discard", 0)
	})

	t.Run("discard usage when no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		out := c.Exec("d")
		assert.NotEmpty(t, out)
		m.AssertNotCalled(t, "Discard", mock.Anything)
	})

	t.Run("knock", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("k 2"))
		m.AssertCalled(t, "Knock", 2)
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("setplayers", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sp 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newMock()
		c := controller.NewThreeThirteenCuiController(m)
		assert.NotEmpty(t, c.Exec("zzz"))
	})
}

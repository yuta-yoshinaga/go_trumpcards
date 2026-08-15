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

func TestKalookiCuiController_Exec(t *testing.T) {
	const mockOutput = `{"phase":0}`

	newMock := func() *mockUsecases.MockKalookiInteractor {
		m := new(mockUsecases.MockKalookiInteractor)
		m.On("GetConfig").Return(domain.DefaultKalookiConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard").Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewKalookiCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultKalookiConfig())
	})

	t.Run("drawstock aliases", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			m := newMock()
			c := controller.NewKalookiCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromStock")
		}
	})

	t.Run("drawdiscard aliases", func(t *testing.T) {
		for _, cmd := range []string{"dd", "drawdiscard"} {
			m := newMock()
			c := controller.NewKalookiCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromDiscard")
		}
	})

	t.Run("meld parses groups", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m 0,1,2 3,4,5"))
		m.AssertCalled(t, "Meld", [][]int{{0, 1, 2}, {3, 4, 5}})
	})

	t.Run("meld usage when no args", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		out := c.Exec("m")
		assert.Contains(t, out, msgUsage("usageMABCDEFOneMeldPerArg"))
		m.AssertNotCalled(t, "Meld", mock.Anything)
	})

	t.Run("layoff parses three ints", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("lo 1 0 2"))
		m.AssertCalled(t, "Layoff", 1, 0, 2)
	})

	t.Run("layoff usage when too few args", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		out := c.Exec("lo 1")
		assert.Contains(t, out, msgUsage("usageLoTargetplayeridxMeldidxCardindex"))
		m.AssertNotCalled(t, "Layoff", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("discard", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 0"))
		m.AssertCalled(t, "Discard", 0)
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("setplayers", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sp 2"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("setthreshold", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("st 41"))
		m.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		m := newMock()
		c := controller.NewKalookiCuiController(m)
		out := c.Exec("zzz")
		assert.NotEmpty(t, out)
	})
}

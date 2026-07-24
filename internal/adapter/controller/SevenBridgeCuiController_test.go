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

func TestSevenBridgeCuiController_Exec(t *testing.T) {
	const mockOutput = `{"phase":0}`

	newMock := func() *mockUsecases.MockSevenBridgeInteractor {
		m := new(mockUsecases.MockSevenBridgeInteractor)
		m.On("GetConfig").Return(domain.DefaultSevenBridgeConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("ClaimPon", mock.Anything).Return(mockOutput)
		m.On("ClaimChi", mock.Anything).Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
		assert.Equal(t, "bye.", c.Exec("quit"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSevenBridgeConfig())
	})

	t.Run("hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("reset word preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("reset"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSevenBridgeConfig())
	})

	t.Run("drawstock aliases", func(t *testing.T) {
		for _, cmd := range []string{"ds", "drawstock"} {
			m := newMock()
			c := controller.NewSevenBridgeCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "DrawFromStock")
		}
	})

	t.Run("pon", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("pon 0,1"))
		m.AssertCalled(t, "ClaimPon", []int{0, 1})
	})

	t.Run("chi", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("chi 2,3"))
		m.AssertCalled(t, "ClaimChi", []int{2, 3})
	})

	t.Run("meld aliases", func(t *testing.T) {
		for _, cmd := range []string{"m 0,1,2", "meld 0,1,2"} {
			m := newMock()
			c := controller.NewSevenBridgeCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "Meld", []int{0, 1, 2})
		}
	})

	t.Run("layoff usage", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		out := c.Exec("lo 0 1")
		assert.Contains(t, out, "Usage: lo")
		m.AssertNotCalled(t, "Layoff", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("layoff ok", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("lo 1,0,2"))
		m.AssertCalled(t, "Layoff", 1, 0, 2)
	})

	t.Run("discard d with idx", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard no args", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		assert.Contains(t, c.Exec("d"), "Card index is required")
	})

	t.Run("discard invalid", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		assert.Contains(t, c.Exec("d abc"), "Invalid card index")
	})

	t.Run("nextround aliases", func(t *testing.T) {
		for _, cmd := range []string{"nr", "nextround"} {
			m := newMock()
			c := controller.NewSevenBridgeCuiController(m)
			assert.Equal(t, mockOutput, c.Exec(cmd))
			m.AssertCalled(t, "NextRound")
		}
	})

	t.Run("setdifficulty valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		want := domain.DefaultSevenBridgeConfig()
		want.CpuDifficulty = domain.SevenBridgeCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", want)
	})

	t.Run("setdifficulty out of range", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		assert.Contains(t, c.Exec("sd 9"), "Invalid CPU difficulty")
	})

	t.Run("setdifficulty no args", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		assert.Contains(t, c.Exec("sd"), "CPU difficulty is required")
	})

	t.Run("setlimit valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 200"))
		want := domain.DefaultSevenBridgeConfig()
		want.PointLimit = 200
		m.AssertCalled(t, "ResetWithConfig", want)
	})

	t.Run("setlimit below min", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		assert.Contains(t, c.Exec("sl 0"), "Invalid point limit")
	})

	t.Run("setlimit no args", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		assert.Contains(t, c.Exec("sl"), "Point limit is required")
	})

	t.Run("log command", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		out := c.Exec("log")
		assert.Equal(t, mockOutput, out)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("log alias l", func(t *testing.T) {
		m := newMock()
		c := controller.NewSevenBridgeCuiController(m)
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		out := c.Exec("bogus")
		assert.NotEmpty(t, out)
	})

	t.Run("empty command", func(t *testing.T) {
		c := controller.NewSevenBridgeCuiController(newMock())
		out := c.Exec("")
		assert.NotEmpty(t, out)
	})
}

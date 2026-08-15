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

func TestBeziqueCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`
	newMock := func() *mockUsecases.MockBeziqueInteractor {
		m := new(mockUsecases.MockBeziqueInteractor)
		m.On("GetConfig").Return(domain.DefaultBeziqueConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("DeclareMeld", mock.Anything).Return(mockOutput)
		m.On("SkipMeld").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit short", func(t *testing.T) {
		c := controller.NewBeziqueCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("r")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultBeziqueConfig())
	})

	t.Run("play with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("p 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Play", 1)
	})

	t.Run("play missing index", func(t *testing.T) {
		c := controller.NewBeziqueCuiController(newMock())
		assert.Contains(t, c.Exec("p"), msgCardIndexRequired())
	})

	t.Run("play invalid index", func(t *testing.T) {
		c := controller.NewBeziqueCuiController(newMock())
		assert.Contains(t, c.Exec("p abc"), msgInvalidCardIndexPrefix())
	})

	t.Run("meld with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("m 0")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "DeclareMeld", 0)
	})

	t.Run("meld missing index", func(t *testing.T) {
		c := controller.NewBeziqueCuiController(newMock())
		assert.Contains(t, c.Exec("meld"), "Meld index is required")
	})

	t.Run("skip short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("s")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "SkipMeld")
	})

	t.Run("next short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("n")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("nextround long", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("nextround")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("sd 2")
		assert.Equal(t, mockOutput, got)
		expected := domain.DefaultBeziqueConfig()
		expected.CpuDifficulty = domain.BeziqueCpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("settarget", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("st 500")
		assert.Equal(t, mockOutput, got)
		expected := domain.DefaultBeziqueConfig()
		expected.TargetScore = 500
		m.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("hint short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("h")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Hint")
	})

	t.Run("log short", func(t *testing.T) {
		m := newMock()
		c := controller.NewBeziqueCuiController(m)
		got := c.Exec("l")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewBeziqueCuiController(newMock())
		got := c.Exec("xyz")
		assert.NotEqual(t, "bye.", got)
		assert.NotEmpty(t, got)
	})
}

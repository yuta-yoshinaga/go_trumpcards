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

func TestRummy500CuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockRummy500Interactor {
		m := new(mockUsecases.MockRummy500Interactor)
		m.On("GetConfig").Return(domain.DefaultRummy500Config())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("DrawFromStock").Return(mockOutput)
		m.On("DrawFromDiscard", mock.Anything).Return(mockOutput)
		m.On("Meld", mock.Anything).Return(mockOutput)
		m.On("Layoff", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything).Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		return m
	}

	t.Run("hint h", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("h"))
		assert.Equal(t, mockOutput, c.Exec("hint"))
		m.AssertCalled(t, "Hint")
	})

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewRummy500CuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("r"))
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultRummy500Config())
	})

	t.Run("drawstock ds", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("ds"))
		m.AssertCalled(t, "DrawFromStock")
	})

	t.Run("drawdiscard dd no args = top", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dd"))
		m.AssertCalled(t, "DrawFromDiscard", -1)
	})

	t.Run("drawdiscard dd with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("dd 2"))
		m.AssertCalled(t, "DrawFromDiscard", 2)
	})

	t.Run("drawdiscard invalid arg", func(t *testing.T) {
		c := controller.NewRummy500CuiController(newMock())
		result := c.Exec("dd xx")
		assert.Contains(t, result, msgInvalidDiscardIndexPrefix())
	})

	t.Run("meld m with indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m 0 1 2"))
		m.AssertCalled(t, "Meld", []int{0, 1, 2})
	})

	t.Run("meld meld with comma-separated indices", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("meld 0,1,2"))
		m.AssertCalled(t, "Meld", []int{0, 1, 2})
	})

	t.Run("meld empty indices is forwarded", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("m"))
		m.AssertCalled(t, "Meld", []int(nil))
	})

	t.Run("layoff lo with 3 ints", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("lo 0 1 2"))
		m.AssertCalled(t, "Layoff", 0, 1, 2)
	})

	t.Run("layoff lo missing args", func(t *testing.T) {
		c := controller.NewRummy500CuiController(newMock())
		result := c.Exec("lo 0 1")
		assert.Contains(t, result, "layoff requires")
	})

	t.Run("discard d with index", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("d 3"))
		m.AssertCalled(t, "Discard", 3)
	})

	t.Run("discard d no args", func(t *testing.T) {
		c := controller.NewRummy500CuiController(newMock())
		result := c.Exec("d")
		assert.Contains(t, result, msgCardIndexRequired())
	})

	t.Run("nextround nr", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("nr"))
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty sd valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sd 2"))
		cfg := domain.DefaultRummy500Config()
		cfg.CpuDifficulty = domain.Rummy500CpuDifficultyHard
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("setdifficulty sd out of range", func(t *testing.T) {
		result := controller.NewRummy500CuiController(newMock()).Exec("sd 9")
		assert.Contains(t, result, "Invalid CPU difficulty")
	})

	t.Run("setlimit sl valid", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("sl 1000"))
		cfg := domain.DefaultRummy500Config()
		cfg.PointLimit = 1000
		m.AssertCalled(t, "ResetWithConfig", cfg)
	})

	t.Run("setlimit sl invalid", func(t *testing.T) {
		result := controller.NewRummy500CuiController(newMock()).Exec("sl 0")
		assert.Contains(t, result, "Invalid point limit")
	})

	t.Run("action log l", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("l"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("action log log", func(t *testing.T) {
		m := newMock()
		c := controller.NewRummy500CuiController(m)
		assert.Equal(t, mockOutput, c.Exec("log"))
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command suggests", func(t *testing.T) {
		result := controller.NewRummy500CuiController(newMock()).Exec("garbage")
		assert.NotEmpty(t, result)
	})

	t.Run("empty command", func(t *testing.T) {
		result := controller.NewRummy500CuiController(newMock()).Exec("")
		assert.NotEmpty(t, result)
	})
}

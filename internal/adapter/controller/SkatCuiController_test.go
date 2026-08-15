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

func TestSkatCuiController_Exec(t *testing.T) {
	mockOutput := `{"phase":0}`

	newMock := func() *mockUsecases.MockSkatInteractor {
		m := new(mockUsecases.MockSkatInteractor)
		m.On("GetConfig").Return(domain.DefaultSkatConfig())
		m.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		m.On("Bid", mock.Anything).Return(mockOutput)
		m.On("PickSkat", mock.Anything).Return(mockOutput)
		m.On("Discard", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("DeclareGame", mock.Anything, mock.Anything).Return(mockOutput)
		m.On("Play", mock.Anything).Return(mockOutput)
		m.On("NextTrick").Return(mockOutput)
		m.On("NextRound").Return(mockOutput)
		m.On("Hint").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}

	t.Run("quit q", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Equal(t, "bye.", c.Exec("q"))
	})

	t.Run("reset r preserves config", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("r")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
		m.AssertCalled(t, "ResetWithConfig", domain.DefaultSkatConfig())
	})

	t.Run("bid accept", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("b 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Bid", true)
	})

	t.Run("bid pass", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("b 0")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Bid", false)
	})

	t.Run("bid invalid", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Contains(t, c.Exec("b abc"), "Invalid bid step")
	})

	t.Run("bid no arg", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Contains(t, c.Exec("b"), "Bid step is required")
	})

	t.Run("pickskat pick up", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("ps 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "PickSkat", true)
	})

	t.Run("pickskat decline", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("pickskat 0")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "PickSkat", false)
	})

	t.Run("discard", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("d 0 2")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Discard", 0, 2)
	})

	t.Run("discard usage", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Contains(t, c.Exec("d 0"), "Usage")
	})

	t.Run("discard invalid first arg", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Contains(t, c.Exec("d abc 1"), msgInvalidCardIndexPrefix())
	})

	t.Run("game suit + trump", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("g 1 1")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "DeclareGame", domain.SkatGameSuit, 1)
	})

	t.Run("game grand", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("g 2")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "DeclareGame", domain.SkatGameGrand, 0)
	})

	t.Run("game usage", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Contains(t, c.Exec("g"), "Usage")
	})

	t.Run("game invalid type", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Contains(t, c.Exec("g 99"), "Invalid game type")
	})

	t.Run("game invalid trump", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Contains(t, c.Exec("g 1 99"), "Invalid trump suit")
	})

	t.Run("play", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("p 3")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Play", 3)
	})

	t.Run("next", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("n")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "NextTrick")
	})

	t.Run("nextround", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("nr")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "NextRound")
	})

	t.Run("setdifficulty", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("sd 2")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
	})

	t.Run("settarget", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("sl 200")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "GetConfig")
	})

	t.Run("hint", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("h")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "Hint")
	})

	t.Run("log", func(t *testing.T) {
		m := newMock()
		c := controller.NewSkatCuiController(m)
		got := c.Exec("log")
		assert.Equal(t, mockOutput, got)
		m.AssertCalled(t, "ActionLog")
	})

	t.Run("unknown command", func(t *testing.T) {
		c := controller.NewSkatCuiController(newMock())
		assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
	})
}

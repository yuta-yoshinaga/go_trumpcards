//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockFourCardPokerInteractor() *usecase.MockFourCardPokerInteractor {
	m := new(usecase.MockFourCardPokerInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0).Return("bet result")
	m.On("Bet", 100, 50).Return("bet with acesup result")
	m.On("Play", 1).Return("play1 result")
	m.On("Play", 2).Return("play2 result")
	m.On("Play", 3).Return("play3 result")
	m.On("Fold").Return("fold result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestFourCardPokerCuiController_Quit(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestFourCardPokerCuiController_Reset(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestFourCardPokerCuiController_Bet(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	t.Run("ante only", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})
	t.Run("ante with aces up", func(t *testing.T) {
		assert.Equal(t, "bet with acesup result", c.Exec("b 100 50"))
	})
	t.Run("bet long form", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestFourCardPokerCuiController_Bet_Errors(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		assert.Contains(t, c.Exec("b"), msgAnteAmountRequired())
	})
	t.Run("invalid amount", func(t *testing.T) {
		assert.Contains(t, c.Exec("b abc"), msgInvalidAnteAmountPrefix())
	})
	t.Run("zero amount", func(t *testing.T) {
		assert.Contains(t, c.Exec("b 0"), msgInvalidAnteAmountPrefix())
	})
	t.Run("negative amount", func(t *testing.T) {
		assert.Contains(t, c.Exec("b -10"), msgInvalidAnteAmountPrefix())
	})
	t.Run("invalid aces up", func(t *testing.T) {
		assert.Contains(t, c.Exec("b 100 abc"), "Invalid Aces Up amount")
	})
}

func TestFourCardPokerCuiController_Play(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	t.Run("default 1x", func(t *testing.T) {
		assert.Equal(t, "play1 result", c.Exec("p"))
		assert.Equal(t, "play1 result", c.Exec("play"))
	})
	t.Run("explicit 1", func(t *testing.T) {
		assert.Equal(t, "play1 result", c.Exec("p 1"))
	})
	t.Run("2x", func(t *testing.T) {
		assert.Equal(t, "play2 result", c.Exec("p 2"))
	})
	t.Run("3x", func(t *testing.T) {
		assert.Equal(t, "play3 result", c.Exec("p 3"))
	})
}

func TestFourCardPokerCuiController_Play_InvalidMultiplier(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	assert.Contains(t, c.Exec("p 4"), "Invalid play multiplier")
	assert.Contains(t, c.Exec("p 0"), "Invalid play multiplier")
	assert.Contains(t, c.Exec("p abc"), "Invalid play multiplier")
}

func TestFourCardPokerCuiController_Fold(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestFourCardPokerCuiController_ActionLog(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestFourCardPokerCuiController_Unknown(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestFourCardPokerCuiController_Empty(t *testing.T) {
	m := newMockFourCardPokerInteractor()
	c := controller.NewFourCardPokerCuiController(m)

	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

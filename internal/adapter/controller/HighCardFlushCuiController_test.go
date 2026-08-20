//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockHighCardFlushInteractor() *usecase.MockHighCardFlushInteractor {
	m := new(usecase.MockHighCardFlushInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0, 0).Return("bet result")
	m.On("Bet", 100, 50, 0).Return("bet with flush bonus result")
	m.On("Bet", 100, 50, 20).Return("bet full sidebets result")
	m.On("Raise", 1).Return("raise 1x result")
	m.On("Raise", 2).Return("raise 2x result")
	m.On("Raise", 3).Return("raise 3x result")
	m.On("Fold").Return("fold result")
	m.On("ActionLog").Return("action log result")
	m.On("Hint").Return("hint result")
	return m
}

func TestHighCardFlushCuiController_Quit(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestHighCardFlushCuiController_Reset(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestHighCardFlushCuiController_Bet(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)

	t.Run("ante only", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})
	t.Run("ante with flush bonus", func(t *testing.T) {
		assert.Equal(t, "bet with flush bonus result", c.Exec("b 100 50"))
	})
	t.Run("ante with both side bets", func(t *testing.T) {
		assert.Equal(t, "bet full sidebets result", c.Exec("b 100 50 20"))
	})
	t.Run("bet long form", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestHighCardFlushCuiController_Bet_Errors(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		assert.Contains(t, c.Exec("b"), msgAnteAmountRequired())
	})
	t.Run("invalid amount", func(t *testing.T) {
		assert.Contains(t, c.Exec("b abc"), msgInvalidAnteAmountPrefix())
	})
	t.Run("zero ante", func(t *testing.T) {
		assert.Contains(t, c.Exec("b 0"), msgInvalidAnteAmountPrefix())
	})
	t.Run("invalid flush bonus", func(t *testing.T) {
		assert.Contains(t, c.Exec("b 100 abc"), msgStem("invalidFlushBonusAmount"))
	})
	t.Run("invalid straight flush bonus", func(t *testing.T) {
		assert.Contains(t, c.Exec("b 100 50 abc"), msgStem("invalidStraightFlushBonusAmount"))
	})
}

func TestHighCardFlushCuiController_Raise(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)

	assert.Equal(t, "raise 1x result", c.Exec("ra 1"))
	assert.Equal(t, "raise 2x result", c.Exec("raise 2"))
	assert.Equal(t, "raise 3x result", c.Exec("ra 3"))
}

func TestHighCardFlushCuiController_Raise_Errors(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)
	assert.Contains(t, c.Exec("ra"), msgStem("raiseMultiplierRequired13"))
	assert.Contains(t, c.Exec("ra 0"), msgStem("invalidRaiseMultiplier"))
	assert.Contains(t, c.Exec("ra 5"), msgStem("invalidRaiseMultiplier"))
	assert.Contains(t, c.Exec("ra xyz"), msgStem("invalidRaiseMultiplier"))
}

func TestHighCardFlushCuiController_Fold(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)
	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestHighCardFlushCuiController_ActionLog(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)
	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestHighCardFlushCuiController_Hint(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)
	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
}

func TestHighCardFlushCuiController_Unknown(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestHighCardFlushCuiController_Empty(t *testing.T) {
	m := newMockHighCardFlushInteractor()
	c := controller.NewHighCardFlushCuiController(m)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

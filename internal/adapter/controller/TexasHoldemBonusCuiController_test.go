//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockTexasHoldemBonusInteractor() *usecase.MockTexasHoldemBonusInteractor {
	m := new(usecase.MockTexasHoldemBonusInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0).Return("bet result")
	m.On("Bet", 100, 10).Return("bet with bonus result")
	m.On("Play").Return("play result")
	m.On("Fold").Return("fold result")
	m.On("Check").Return("check result")
	m.On("Raise").Return("raise result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestTexasHoldemBonusCuiController_Quit(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestTexasHoldemBonusCuiController_Reset(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestTexasHoldemBonusCuiController_Bet(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)

	t.Run("ante only", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("b 100"))
	})
	t.Run("ante with bonus", func(t *testing.T) {
		assert.Equal(t, "bet with bonus result", c.Exec("b 100 10"))
	})
	t.Run("bet long form", func(t *testing.T) {
		assert.Equal(t, "bet result", c.Exec("bet 100"))
	})
}

func TestTexasHoldemBonusCuiController_Bet_Errors(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		result := c.Exec("b")
		assert.Contains(t, result, msgAnteAmountRequired())
	})
	t.Run("invalid amount", func(t *testing.T) {
		result := c.Exec("b abc")
		assert.Contains(t, result, msgInvalidAnteAmountPrefix())
	})
	t.Run("zero amount", func(t *testing.T) {
		result := c.Exec("b 0")
		assert.Contains(t, result, msgInvalidAnteAmountPrefix())
	})
	t.Run("invalid bonus", func(t *testing.T) {
		result := c.Exec("b 100 abc")
		assert.Contains(t, result, msgStem("invalidBonusAmount"))
	})
}

func TestTexasHoldemBonusCuiController_Play(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	assert.Equal(t, "play result", c.Exec("p"))
	assert.Equal(t, "play result", c.Exec("play"))
}

func TestTexasHoldemBonusCuiController_Fold(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestTexasHoldemBonusCuiController_Check(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	assert.Equal(t, "check result", c.Exec("c"))
	assert.Equal(t, "check result", c.Exec("check"))
}

func TestTexasHoldemBonusCuiController_Raise(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	assert.Equal(t, "raise result", c.Exec("ra"))
	assert.Equal(t, "raise result", c.Exec("raise"))
}

func TestTexasHoldemBonusCuiController_ActionLog(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestTexasHoldemBonusCuiController_Unknown(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestTexasHoldemBonusCuiController_Empty(t *testing.T) {
	m := newMockTexasHoldemBonusInteractor()
	c := controller.NewTexasHoldemBonusCuiController(m)
	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

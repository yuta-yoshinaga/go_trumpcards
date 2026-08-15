//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCasinoHoldemInteractor() *usecase.MockCasinoHoldemInteractor {
	m := new(usecase.MockCasinoHoldemInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0).Return("bet result")
	m.On("Bet", 100, 10).Return("bet with bonus result")
	m.On("Call").Return("call result")
	m.On("Fold").Return("fold result")
	m.On("ActionLog").Return("action log result")
	m.On("Hint").Return("hint result")
	return m
}

func TestCasinoHoldemCuiController_Quit(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCasinoHoldemCuiController_Reset(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestCasinoHoldemCuiController_Bet(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)

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

func TestCasinoHoldemCuiController_Bet_Errors(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)

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

func TestCasinoHoldemCuiController_Call(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)
	assert.Equal(t, "call result", c.Exec("c"))
	assert.Equal(t, "call result", c.Exec("call"))
}

func TestCasinoHoldemCuiController_Fold(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)
	assert.Equal(t, "fold result", c.Exec("f"))
	assert.Equal(t, "fold result", c.Exec("fold"))
}

func TestCasinoHoldemCuiController_ActionLog(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)
	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestCasinoHoldemCuiController_Unknown(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestCasinoHoldemCuiController_Empty(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)
	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

func TestCasinoHoldemCuiController_Hint(t *testing.T) {
	m := newMockCasinoHoldemInteractor()
	c := controller.NewCasinoHoldemCuiController(m)
	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
	m.AssertCalled(t, "Hint")
}

//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockBaccaratInteractor() *usecase.MockBaccaratInteractor {
	m := new(usecase.MockBaccaratInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, 0, 0, 0).Return("bet player result")
	m.On("Bet", 100, 1, 0, 0).Return("bet banker result")
	m.On("Bet", 100, 2, 0, 0).Return("bet tie result")
	m.On("ActionLog").Return("action log result")
	m.On("ClearHistory").Return("clear history result")
	return m
}

func TestBaccaratCuiController_Quit(t *testing.T) {
	m := newMockBaccaratInteractor()
	c := controller.NewBaccaratCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestBaccaratCuiController_Reset(t *testing.T) {
	m := newMockBaccaratInteractor()
	c := controller.NewBaccaratCuiController(m)

	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestBaccaratCuiController_Bet(t *testing.T) {
	m := newMockBaccaratInteractor()
	c := controller.NewBaccaratCuiController(m)

	t.Run("player bet", func(t *testing.T) {
		assert.Equal(t, "bet player result", c.Exec("b 100 0"))
	})

	t.Run("banker bet", func(t *testing.T) {
		assert.Equal(t, "bet banker result", c.Exec("b 100 1"))
	})

	t.Run("tie bet", func(t *testing.T) {
		assert.Equal(t, "bet tie result", c.Exec("b 100 2"))
	})

	t.Run("bet long form", func(t *testing.T) {
		assert.Equal(t, "bet player result", c.Exec("bet 100 0"))
	})
}

func TestBaccaratCuiController_Bet_Errors(t *testing.T) {
	m := newMockBaccaratInteractor()
	c := controller.NewBaccaratCuiController(m)

	t.Run("missing args", func(t *testing.T) {
		result := c.Exec("b")
		assert.Contains(t, result, msgBetAmountRequired())
	})

	t.Run("missing bet type", func(t *testing.T) {
		result := c.Exec("b 100")
		assert.Contains(t, result, "Bet type is required")
	})

	t.Run("invalid amount", func(t *testing.T) {
		result := c.Exec("b abc 0")
		assert.Contains(t, result, msgInvalidBetAmountPrefix())
	})

	t.Run("zero amount", func(t *testing.T) {
		result := c.Exec("b 0 0")
		assert.Contains(t, result, msgInvalidBetAmountPrefix())
	})

	t.Run("negative amount", func(t *testing.T) {
		result := c.Exec("b -10 0")
		assert.Contains(t, result, msgInvalidBetAmountPrefix())
	})

	t.Run("invalid bet type text", func(t *testing.T) {
		result := c.Exec("b 100 abc")
		assert.Contains(t, result, "Invalid bet type")
	})

	t.Run("bet type out of range high", func(t *testing.T) {
		result := c.Exec("b 100 3")
		assert.Contains(t, result, "Invalid bet type")
	})

	t.Run("bet type out of range low", func(t *testing.T) {
		result := c.Exec("b 100 -1")
		assert.Contains(t, result, "Invalid bet type")
	})
}

func TestBaccaratCuiController_ActionLog(t *testing.T) {
	m := newMockBaccaratInteractor()
	c := controller.NewBaccaratCuiController(m)

	assert.Equal(t, "action log result", c.Exec("log"))
	assert.Equal(t, "action log result", c.Exec("l"))
}

func TestBaccaratCuiController_ClearHistory(t *testing.T) {
	m := newMockBaccaratInteractor()
	c := controller.NewBaccaratCuiController(m)

	assert.Equal(t, "clear history result", c.Exec("ch"))
	assert.Equal(t, "clear history result", c.Exec("clearhistory"))
}

func TestBaccaratCuiController_Unknown(t *testing.T) {
	m := newMockBaccaratInteractor()
	c := controller.NewBaccaratCuiController(m)

	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestBaccaratCuiController_Empty(t *testing.T) {
	m := newMockBaccaratInteractor()
	c := controller.NewBaccaratCuiController(m)

	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

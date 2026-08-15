//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockDragonTigerInteractor() *usecase.MockDragonTigerInteractor {
	m := new(usecase.MockDragonTigerInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100, domain.DragonTigerBetDragon).Return("bet dragon")
	m.On("Bet", 100, domain.DragonTigerBetTiger).Return("bet tiger")
	m.On("Bet", 100, domain.DragonTigerBetTie).Return("bet tie")
	m.On("ClearHistory").Return("history cleared")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestDragonTigerCuiController_Quit(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestDragonTigerCuiController_Reset(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestDragonTigerCuiController_Bet_Dragon(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "bet dragon", c.Exec("b 100 d"))
	assert.Equal(t, "bet dragon", c.Exec("bet 100 dragon"))
	assert.Equal(t, "bet dragon", c.Exec("b 100 0"))
}

func TestDragonTigerCuiController_Bet_Tiger(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "bet tiger", c.Exec("b 100 t"))
	assert.Equal(t, "bet tiger", c.Exec("b 100 tiger"))
	assert.Equal(t, "bet tiger", c.Exec("b 100 1"))
}

func TestDragonTigerCuiController_Bet_Tie(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "bet tie", c.Exec("b 100 e"))
	assert.Equal(t, "bet tie", c.Exec("b 100 tie"))
	assert.Equal(t, "bet tie", c.Exec("b 100 2"))
}

func TestDragonTigerCuiController_Bet_Errors(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Contains(t, c.Exec("b"), "Bet amount and type")
	assert.Contains(t, c.Exec("b 100"), "Bet amount and type")
	assert.Contains(t, c.Exec("b abc d"), msgInvalidBetAmountPrefix())
	assert.Contains(t, c.Exec("b 100 x"), "Invalid bet type")
}

func TestDragonTigerCuiController_ClearHistory(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "history cleared", c.Exec("clear"))
}

func TestDragonTigerCuiController_ActionLog(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestDragonTigerCuiController_Unknown(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestDragonTigerCuiController_Empty(t *testing.T) {
	c := controller.NewDragonTigerCuiController(newMockDragonTigerInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

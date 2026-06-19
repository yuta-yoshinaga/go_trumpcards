//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockFaroInteractor() *usecase.MockFaroInteractor {
	m := new(usecase.MockFaroInteractor)
	m.On("Reset").Return("reset result")
	m.On("NextRound").Return("next result")
	m.On("PlaceBet", 7, 100, false).Return("bet result")
	m.On("PlaceBet", 7, 100, true).Return("bet copper result")
	m.On("ClearBet", 3).Return("clearbet result")
	m.On("ClearAll").Return("clearall result")
	m.On("DealTurn").Return("deal result")
	m.On("Call", []int{3, 9, 12}).Return("call result")
	m.On("Call", []int(nil)).Return("call skip result")
	m.On("ActionLog").Return("log result")
	return m
}

func TestFaroCuiController_Quit(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestFaroCuiController_Reset(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestFaroCuiController_Bet(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "bet result", c.Exec("b 7 100"))
	assert.Equal(t, "bet copper result", c.Exec("b 7 100 c"))
}

func TestFaroCuiController_Bet_Errors(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.NotEmpty(t, c.Exec("b"))
	assert.NotEmpty(t, c.Exec("b 7"))
	assert.NotEmpty(t, c.Exec("b x 100"))
	assert.NotEmpty(t, c.Exec("b 7 xyz"))
}

func TestFaroCuiController_ClearBet(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "clearbet result", c.Exec("cb 3"))
	assert.NotEmpty(t, c.Exec("cb"))
	assert.NotEmpty(t, c.Exec("cb x"))
}

func TestFaroCuiController_ClearAll(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "clearall result", c.Exec("ca"))
}

func TestFaroCuiController_Deal(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "deal result", c.Exec("d"))
	assert.Equal(t, "deal result", c.Exec("deal"))
}

func TestFaroCuiController_Call(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "call result", c.Exec("call 3 9 12"))
	assert.Equal(t, "call skip result", c.Exec("call"))
	assert.NotEmpty(t, c.Exec("call 3 x 12"))
}

func TestFaroCuiController_Next(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "next result", c.Exec("n"))
	assert.Equal(t, "next result", c.Exec("next"))
}

func TestFaroCuiController_ActionLog(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Equal(t, "log result", c.Exec("log"))
}

func TestFaroCuiController_Unknown(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestFaroCuiController_Empty(t *testing.T) {
	c := controller.NewFaroCuiController(newMockFaroInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

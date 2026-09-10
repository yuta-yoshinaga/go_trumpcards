//go:build test
// +build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func TestBassetCuiController_ParoliCommand(t *testing.T) {
	m := new(uc.MockBassetInteractor)
	m.On("PressParoli").Return("ok")
	c := controller.NewBassetCuiController(m)
	assert.Equal(t, "ok", c.Exec("paroli"))
	m.AssertCalled(t, "PressParoli")
}
func TestBassetCuiController_BetCommand(t *testing.T) {
	m := new(uc.MockBassetInteractor)
	m.On("PlaceBet", 7, 10).Return("ok")
	c := controller.NewBassetCuiController(m)
	assert.Equal(t, "ok", c.Exec("bet 7 10"))
	m.AssertCalled(t, "PlaceBet", 7, 10)
}

func newBassetMockInteractor() *uc.MockBassetInteractor {
	m := new(uc.MockBassetInteractor)
	m.On("Reset").Return("reset result")
	m.On("NextRound").Return("next result")
	m.On("PlaceBet", 7, 10).Return("bet result")
	m.On("DealTurn").Return("deal result")
	m.On("TakeWinnings").Return("take result")
	m.On("PressParoli").Return("paroli result")
	m.On("ActionLog").Return("log result")
	return m
}

func TestBassetCuiController_Commands(t *testing.T) {
	c := controller.NewBassetCuiController(newBassetMockInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
	assert.Equal(t, "bet result", c.Exec("b 7 10"))
	assert.Equal(t, "bet result", c.Exec("bet 7 10"))
	assert.Equal(t, "deal result", c.Exec("d"))
	assert.Equal(t, "deal result", c.Exec("deal"))
	assert.Equal(t, "take result", c.Exec("take"))
	assert.Equal(t, "paroli result", c.Exec("paroli"))
	assert.Equal(t, "next result", c.Exec("n"))
	assert.Equal(t, "next result", c.Exec("next"))
	assert.Equal(t, "log result", c.Exec("log"))
}

func TestBassetCuiController_InvalidCommands(t *testing.T) {
	c := controller.NewBassetCuiController(newBassetMockInteractor())
	for _, command := range []string{"b", "b 7", "b x 10", "b 7 x"} {
		assert.NotEmpty(t, c.Exec(command))
	}
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
	assert.Contains(t, c.Exec("unknown"), "コマンドが不明です")
}

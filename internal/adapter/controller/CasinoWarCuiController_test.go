//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCasinoWarInteractor() *usecase.MockCasinoWarInteractor {
	m := new(usecase.MockCasinoWarInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("Surrender").Return("surrender result")
	m.On("War").Return("war result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestCasinoWarCuiController_Quit(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCasinoWarCuiController_Reset(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestCasinoWarCuiController_Bet(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "bet result", c.Exec("b 100"))
	assert.Equal(t, "bet result", c.Exec("bet 100"))
}

func TestCasinoWarCuiController_Bet_Errors(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Contains(t, c.Exec("b"), msgBetAmountRequired())
	assert.Contains(t, c.Exec("b abc"), msgInvalidBetAmountPrefix())
	assert.Contains(t, c.Exec("b 0"), msgInvalidBetAmountPrefix())
}

func TestCasinoWarCuiController_Surrender(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "surrender result", c.Exec("surrender"))
}

func TestCasinoWarCuiController_War(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "war result", c.Exec("war"))
}

func TestCasinoWarCuiController_ActionLog(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestCasinoWarCuiController_Unknown(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestCasinoWarCuiController_Empty(t *testing.T) {
	c := controller.NewCasinoWarCuiController(newMockCasinoWarInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockOichoKabuInteractor() *usecase.MockOichoKabuInteractor {
	m := new(usecase.MockOichoKabuInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("Draw").Return("draw result")
	m.On("Stand").Return("stand result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestOichoKabuCuiController_Quit(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestOichoKabuCuiController_Reset(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestOichoKabuCuiController_Bet(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Equal(t, "bet result", c.Exec("b 100"))
	assert.Equal(t, "bet result", c.Exec("bet 100"))
}

func TestOichoKabuCuiController_Bet_Errors(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Contains(t, c.Exec("b"), msgBetAmountRequired())
	assert.Contains(t, c.Exec("b abc"), msgInvalidBetAmountPrefix())
	assert.Contains(t, c.Exec("b 0"), msgInvalidBetAmountPrefix())
}

func TestOichoKabuCuiController_Draw(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Equal(t, "draw result", c.Exec("draw"))
	assert.Equal(t, "draw result", c.Exec("d"))
}

func TestOichoKabuCuiController_Stand(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Equal(t, "stand result", c.Exec("stand"))
	assert.Equal(t, "stand result", c.Exec("s"))
}

func TestOichoKabuCuiController_ActionLog(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestOichoKabuCuiController_Unknown(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestOichoKabuCuiController_Empty(t *testing.T) {
	c := controller.NewOichoKabuCuiController(newMockOichoKabuInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

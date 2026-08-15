//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockRedDogInteractor() *usecase.MockRedDogInteractor {
	m := new(usecase.MockRedDogInteractor)
	m.On("Reset").Return("reset result")
	m.On("Bet", 100).Return("bet result")
	m.On("Raise", 50).Return("raise result")
	m.On("Stay").Return("stay result")
	m.On("Hint").Return("hint result")
	m.On("ActionLog").Return("action log result")
	return m
}

func TestRedDogCuiController_Quit(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestRedDogCuiController_Reset(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Equal(t, "reset result", c.Exec("r"))
	assert.Equal(t, "reset result", c.Exec("reset"))
}

func TestRedDogCuiController_Bet(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Equal(t, "bet result", c.Exec("b 100"))
	assert.Equal(t, "bet result", c.Exec("bet 100"))
}

func TestRedDogCuiController_Bet_Errors(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Contains(t, c.Exec("b"), msgBetAmountRequired())
	assert.Contains(t, c.Exec("b abc"), msgInvalidBetAmountPrefix())
	assert.Contains(t, c.Exec("b 0"), msgInvalidBetAmountPrefix())
}

func TestRedDogCuiController_Raise(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Equal(t, "raise result", c.Exec("raise 50"))
}

func TestRedDogCuiController_Raise_Errors(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Contains(t, c.Exec("raise"), "Raise amount is required")
	assert.Contains(t, c.Exec("raise abc"), "Invalid raise amount")
}

func TestRedDogCuiController_Stay(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Equal(t, "stay result", c.Exec("s"))
	assert.Equal(t, "stay result", c.Exec("stay"))
}

func TestRedDogCuiController_Hint(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Equal(t, "hint result", c.Exec("h"))
	assert.Equal(t, "hint result", c.Exec("hint"))
}

func TestRedDogCuiController_ActionLog(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Equal(t, "action log result", c.Exec("log"))
}

func TestRedDogCuiController_Unknown(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestRedDogCuiController_Empty(t *testing.T) {
	c := controller.NewRedDogCuiController(newMockRedDogInteractor())
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

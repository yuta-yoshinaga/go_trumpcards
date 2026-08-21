package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockAlaskaInteractor() *mockusecase.MockAlaskaInteractor {
	return new(mockusecase.MockAlaskaInteractor)
}

func TestAlaskaCuiControllerQuit(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestAlaskaCuiControllerReset(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestAlaskaCuiControllerGiveUp(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestAlaskaCuiControllerHint(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestAlaskaCuiControllerAutoComplete(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestAlaskaCuiControllerActionLog(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestAlaskaCuiControllerUndo(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestAlaskaCuiControllerMoveShorthand(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("MoveTableauToTableau", 0, -1, 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 3"))
}

func TestAlaskaCuiControllerMoveShorthandPrompt(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	result := c.Exec("m 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestAlaskaCuiControllerMovePrompt(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	result := c.Exec("m")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestAlaskaCuiControllerMoveTableauToFoundation(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("MoveTableauToFoundation", 2).Return("move_f_output")
	assert.Equal(t, "move_f_output", c.Exec("m t 2 f"))
}

func TestAlaskaCuiControllerMoveTableauToTableau(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)
	ri.On("MoveTableauToTableau", 0, 2, 4).Return("move_tt_output")
	assert.Equal(t, "move_tt_output", c.Exec("m t 0 2 t 4"))
}

func TestAlaskaCuiControllerMoveTableauPrompts(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)

	result := c.Exec("m t")
	assert.Contains(t, result, cuiutil.PromptPrefix)

	result = c.Exec("m t 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)

	result = c.Exec("m t 0 2 t")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestAlaskaCuiControllerMoveInvalidFrom(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)

	result := c.Exec("m w")
	assert.NotEmpty(t, result)
}

func TestAlaskaCuiControllerMoveUsage(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)

	result := c.Exec("m t 0 2 x")
	assert.NotEmpty(t, result)
}

func TestAlaskaCuiControllerMoveInvalidCol(t *testing.T) {
	ri := newMockAlaskaInteractor()
	c := NewAlaskaCuiController(ri)

	result := c.Exec("m t abc")
	assert.NotEmpty(t, result)

	result = c.Exec("m 0 abc")
	assert.NotEmpty(t, result)
}

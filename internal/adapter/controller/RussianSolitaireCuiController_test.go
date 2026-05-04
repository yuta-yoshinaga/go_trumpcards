package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockRussianSolitaireInteractor() *mockusecase.MockRussianSolitaireInteractor {
	return new(mockusecase.MockRussianSolitaireInteractor)
}

func TestRussianSolitaireCuiControllerQuit(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestRussianSolitaireCuiControllerReset(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestRussianSolitaireCuiControllerGiveUp(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestRussianSolitaireCuiControllerHint(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestRussianSolitaireCuiControllerAutoComplete(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestRussianSolitaireCuiControllerActionLog(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestRussianSolitaireCuiControllerUndo(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestRussianSolitaireCuiControllerMoveShorthand(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("MoveTableauToTableau", 0, -1, 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 3"))
}

func TestRussianSolitaireCuiControllerMoveShorthandPrompt(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	result := c.Exec("m 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestRussianSolitaireCuiControllerMovePrompt(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	result := c.Exec("m")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestRussianSolitaireCuiControllerMoveTableauToFoundation(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("MoveTableauToFoundation", 2).Return("move_f_output")
	assert.Equal(t, "move_f_output", c.Exec("m t 2 f"))
}

func TestRussianSolitaireCuiControllerMoveTableauToTableau(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)
	ri.On("MoveTableauToTableau", 0, 2, 4).Return("move_tt_output")
	assert.Equal(t, "move_tt_output", c.Exec("m t 0 2 t 4"))
}

func TestRussianSolitaireCuiControllerMoveTableauPrompts(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)

	result := c.Exec("m t")
	assert.Contains(t, result, cuiutil.PromptPrefix)

	result = c.Exec("m t 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)

	result = c.Exec("m t 0 2 t")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestRussianSolitaireCuiControllerMoveInvalidFrom(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)

	result := c.Exec("m w")
	assert.NotEmpty(t, result)
}

func TestRussianSolitaireCuiControllerMoveUsage(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)

	result := c.Exec("m t 0 2 x")
	assert.NotEmpty(t, result)
}

func TestRussianSolitaireCuiControllerMoveInvalidCol(t *testing.T) {
	ri := newMockRussianSolitaireInteractor()
	c := NewRussianSolitaireCuiController(ri)

	result := c.Exec("m t abc")
	assert.NotEmpty(t, result)

	result = c.Exec("m 0 abc")
	assert.NotEmpty(t, result)
}

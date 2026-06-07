package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockEasthavenInteractor() *mockusecase.MockEasthavenInteractor {
	return new(mockusecase.MockEasthavenInteractor)
}

func TestEasthavenCuiControllerQuit(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestEasthavenCuiControllerReset(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestEasthavenCuiControllerDeal(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("Deal").Return("deal_output")
	assert.Equal(t, "deal_output", c.Exec("d"))
	assert.Equal(t, "deal_output", c.Exec("deal"))
}

func TestEasthavenCuiControllerGiveUp(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestEasthavenCuiControllerHint(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestEasthavenCuiControllerAutoComplete(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestEasthavenCuiControllerActionLog(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestEasthavenCuiControllerUndo(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestEasthavenCuiControllerMoveShorthand(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("MoveTableauToTableau", 0, -1, 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 3"))
}

func TestEasthavenCuiControllerMoveShorthandPrompt(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	assert.Contains(t, c.Exec("m 0"), cuiutil.PromptPrefix)
}

func TestEasthavenCuiControllerMovePrompt(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	assert.Contains(t, c.Exec("m"), cuiutil.PromptPrefix)
}

func TestEasthavenCuiControllerMoveTableauToFoundation(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("MoveTableauToFoundation", 2).Return("move_f_output")
	assert.Equal(t, "move_f_output", c.Exec("m t 2 f"))
}

func TestEasthavenCuiControllerMoveTableauToTableau(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	ei.On("MoveTableauToTableau", 0, 2, 4).Return("move_tt_output")
	assert.Equal(t, "move_tt_output", c.Exec("m t 0 2 t 4"))
}

func TestEasthavenCuiControllerMoveTableauPrompts(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	assert.Contains(t, c.Exec("m t"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m t 0"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m t 0 2 t"), cuiutil.PromptPrefix)
}

func TestEasthavenCuiControllerMoveInvalidFrom(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	assert.NotEmpty(t, c.Exec("m w"))
}

func TestEasthavenCuiControllerMoveUsage(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	assert.NotEmpty(t, c.Exec("m t 0 2 x"))
}

func TestEasthavenCuiControllerMoveInvalidCol(t *testing.T) {
	ei := newMockEasthavenInteractor()
	c := NewEasthavenCuiController(ei)
	assert.NotEmpty(t, c.Exec("m t abc"))
	assert.NotEmpty(t, c.Exec("m 0 abc"))
}

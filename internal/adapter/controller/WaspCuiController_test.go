package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockWaspInteractor() *mockusecase.MockWaspInteractor {
	return new(mockusecase.MockWaspInteractor)
}

func TestWaspCuiControllerQuit(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestWaspCuiControllerReset(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestWaspCuiControllerDeal(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("Deal").Return("deal_output")
	assert.Equal(t, "deal_output", c.Exec("d"))
	assert.Equal(t, "deal_output", c.Exec("deal"))
}

func TestWaspCuiControllerGiveUp(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestWaspCuiControllerHint(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestWaspCuiControllerAutoComplete(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestWaspCuiControllerActionLog(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestWaspCuiControllerUndo(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestWaspCuiControllerLegal(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("LegalMoves", 2).Return("legal_output")
	assert.Equal(t, "legal_output", c.Exec("legal 2"))
}

func TestWaspCuiControllerLegalPrompt(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	result := c.Exec("legal")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestWaspCuiControllerLegalInvalidCol(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	assert.NotEmpty(t, c.Exec("legal abc"))
}

func TestWaspCuiControllerMoveShorthandTopCard(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("MoveTableauToTableau", 0, -1, 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 3"))
}

func TestWaspCuiControllerMoveShorthandWithIdx(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("MoveTableauToTableau", 0, 2, 4).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 2 4"))
}

func TestWaspCuiControllerMoveShorthandPrompt(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	result := c.Exec("m 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestWaspCuiControllerMovePrompt(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	result := c.Exec("m")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestWaspCuiControllerMoveTableauToTableau(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)
	si.On("MoveTableauToTableau", 0, 2, 4).Return("move_tt_output")
	assert.Equal(t, "move_tt_output", c.Exec("m t 0 2 t 4"))
}

func TestWaspCuiControllerMoveTableauPrompts(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)

	assert.Contains(t, c.Exec("m t"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m t 0"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m t 0 2 t"), cuiutil.PromptPrefix)
}

func TestWaspCuiControllerMoveInvalidFrom(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)

	result := c.Exec("m w")
	assert.NotEmpty(t, result)
}

func TestWaspCuiControllerMoveUsage(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)

	result := c.Exec("m t 0 2 x")
	assert.NotEmpty(t, result)
}

func TestWaspCuiControllerMoveInvalidCol(t *testing.T) {
	si := newMockWaspInteractor()
	c := NewWaspCuiController(si)

	assert.NotEmpty(t, c.Exec("m t abc"))
	assert.NotEmpty(t, c.Exec("m 0 abc"))
	assert.NotEmpty(t, c.Exec("m 0 1 abc"))
	assert.NotEmpty(t, c.Exec("m t 0 abc t 4"))
	assert.NotEmpty(t, c.Exec("m t 0 1 t abc"))
}

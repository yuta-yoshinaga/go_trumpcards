package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockScorpionInteractor() *mockusecase.MockScorpionInteractor {
	return new(mockusecase.MockScorpionInteractor)
}

func TestScorpionCuiControllerQuit(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestScorpionCuiControllerReset(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestScorpionCuiControllerDeal(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("Deal").Return("deal_output")
	assert.Equal(t, "deal_output", c.Exec("d"))
	assert.Equal(t, "deal_output", c.Exec("deal"))
}

func TestScorpionCuiControllerLegal(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("LegalMoves", 2).Return("legal_output")
	assert.Equal(t, "legal_output", c.Exec("legal 2"))
}

func TestScorpionCuiControllerLegalPrompt(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	result := c.Exec("legal")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestScorpionCuiControllerLegalInvalidCol(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	assert.NotEmpty(t, c.Exec("legal abc"))
}

func TestScorpionCuiControllerGiveUp(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestScorpionCuiControllerHint(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestScorpionCuiControllerAutoComplete(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestScorpionCuiControllerActionLog(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestScorpionCuiControllerUndo(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestScorpionCuiControllerMoveShorthandTopCard(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("MoveTableauToTableau", 0, -1, 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 3"))
}

func TestScorpionCuiControllerMoveShorthandWithIdx(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("MoveTableauToTableau", 0, 2, 4).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 2 4"))
}

func TestScorpionCuiControllerMoveShorthandPrompt(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	result := c.Exec("m 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestScorpionCuiControllerMovePrompt(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	result := c.Exec("m")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestScorpionCuiControllerMoveTableauToTableau(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("MoveTableauToTableau", 0, 2, 4).Return("move_tt_output")
	assert.Equal(t, "move_tt_output", c.Exec("m t 0 2 t 4"))
}

func TestScorpionCuiControllerMoveTableauPrompts(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)

	assert.Contains(t, c.Exec("m t"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m t 0"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m t 0 2 t"), cuiutil.PromptPrefix)
}

func TestScorpionCuiControllerMoveInvalidFrom(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)

	result := c.Exec("m w")
	assert.NotEmpty(t, result)
}

func TestScorpionCuiControllerMoveUsage(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)

	result := c.Exec("m t 0 2 x")
	assert.NotEmpty(t, result)
}

func TestScorpionCuiControllerMoveInvalidCol(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)

	assert.NotEmpty(t, c.Exec("m t abc"))
	assert.NotEmpty(t, c.Exec("m 0 abc"))
	assert.NotEmpty(t, c.Exec("m 0 1 abc"))
	assert.NotEmpty(t, c.Exec("m t 0 abc t 4"))
	assert.NotEmpty(t, c.Exec("m t 0 1 t abc"))
}

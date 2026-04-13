package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockYukonInteractor() *mockusecase.MockYukonInteractor {
	return new(mockusecase.MockYukonInteractor)
}

func TestYukonCuiControllerQuit(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestYukonCuiControllerReset(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestYukonCuiControllerGiveUp(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestYukonCuiControllerHint(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestYukonCuiControllerAutoComplete(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestYukonCuiControllerActionLog(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestYukonCuiControllerUndo(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestYukonCuiControllerMoveShorthand(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("MoveTableauToTableau", 0, -1, 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 3"))
}

func TestYukonCuiControllerMoveShorthandPrompt(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	result := c.Exec("m 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestYukonCuiControllerMovePrompt(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	result := c.Exec("m")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestYukonCuiControllerMoveTableauToFoundation(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("MoveTableauToFoundation", 2).Return("move_f_output")
	assert.Equal(t, "move_f_output", c.Exec("m t 2 f"))
}

func TestYukonCuiControllerMoveTableauToTableau(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)
	yi.On("MoveTableauToTableau", 0, 2, 4).Return("move_tt_output")
	assert.Equal(t, "move_tt_output", c.Exec("m t 0 2 t 4"))
}

func TestYukonCuiControllerMoveTableauPrompts(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)

	// "m t" -> prompt for from column
	result := c.Exec("m t")
	assert.Contains(t, result, cuiutil.PromptPrefix)

	// "m t 0" -> prompt for to zone
	result = c.Exec("m t 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)

	// "m t 0 2 t" -> prompt for to column
	result = c.Exec("m t 0 2 t")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestYukonCuiControllerMoveInvalidFrom(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)

	// "m w" -> invalid (no waste in Yukon)
	result := c.Exec("m w")
	assert.NotEmpty(t, result)
}

func TestYukonCuiControllerMoveUsage(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)

	// "m t 0 2 x" -> usage message
	result := c.Exec("m t 0 2 x")
	assert.NotEmpty(t, result)
}

func TestYukonCuiControllerMoveInvalidCol(t *testing.T) {
	yi := newMockYukonInteractor()
	c := NewYukonCuiController(yi)

	result := c.Exec("m t abc")
	assert.NotEmpty(t, result)

	result = c.Exec("m 0 abc")
	assert.NotEmpty(t, result)
}

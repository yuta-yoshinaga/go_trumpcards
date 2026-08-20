package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockCrescentInteractor() *mockusecase.MockCrescentInteractor {
	return new(mockusecase.MockCrescentInteractor)
}

func TestCrescentCuiControllerQuit(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCrescentCuiControllerReset(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("Reset").Return("reset_output")
	// "r" and "reset" both route to reset via the shared cui helper.
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestCrescentCuiControllerRedeal(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("Redeal").Return("redeal_output")
	// "rd" is the dedicated redeal alias because "r" is reserved for reset.
	assert.Equal(t, "redeal_output", c.Exec("rd"))
	assert.Equal(t, "redeal_output", c.Exec("redeal"))
}

func TestCrescentCuiControllerGiveUp(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestCrescentCuiControllerHint(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestCrescentCuiControllerAutoComplete(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestCrescentCuiControllerActionLog(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestCrescentCuiControllerUndo(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestCrescentCuiControllerMoveShorthand(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("MoveTableauToTableau", 3, 5).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 3 5"))
}

func TestCrescentCuiControllerMoveTableauToTableauExplicit(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("MoveTableauToTableau", 3, 5).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 3 t 5"))
}

func TestCrescentCuiControllerMoveTableauToFoundation(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	ci.On("MoveTableauToFoundation", 1, 4).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 1 f 4"))
}

func TestCrescentCuiControllerMovePrompts(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("m")))
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t")))
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t 3")))
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t 3 t")))
	assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t 3 f")))
}

// codecov: the default arm of the move parser -- an `m` shape the game does
// not have answers with the usage line.
func TestCrescentCuiControllerMoveUsage(t *testing.T) {
	ci := newMockCrescentInteractor()
	c := NewCrescentCuiController(ci)
	out := c.Exec("m t 0 x 1")
	body, isErr := i18n.StripErrorPrefix(out)
	assert.True(t, isErr, "a usage line means the move did not happen")
	assert.Equal(t, i18n.T("crescent.moveUsage"), body)
}

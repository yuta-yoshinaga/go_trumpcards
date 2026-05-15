package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCruelInteractor() *mockusecase.MockCruelInteractor {
	return new(mockusecase.MockCruelInteractor)
}

func TestCruelCuiControllerQuit(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCruelCuiControllerReset(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestCruelCuiControllerShift(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("Shift").Return("shift_output")
	assert.Equal(t, "shift_output", c.Exec("s"))
	assert.Equal(t, "shift_output", c.Exec("shift"))
}

func TestCruelCuiControllerGiveUp(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestCruelCuiControllerHint(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestCruelCuiControllerAutoComplete(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestCruelCuiControllerActionLog(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestCruelCuiControllerUndo(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestCruelCuiControllerMoveTableauToTableau(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("MoveTableauToTableau", 0, 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 3"))
}

func TestCruelCuiControllerMoveTableauToFoundation(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	ci.On("MoveTableauToFoundation", 5).Return("move_f_output")
	assert.Equal(t, "move_f_output", c.Exec("m 5 f"))
}

func TestCruelCuiControllerMovePromptSource(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	got := c.Exec("m")
	// The first column prompt should mention the source column.
	assert.Contains(t, got, "m ")
}

func TestCruelCuiControllerMovePromptDestination(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	got := c.Exec("m 0")
	assert.Contains(t, got, "m 0 ")
}

func TestCruelCuiControllerMoveInvalidSourceColumn(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	got := c.Exec("m abc 3")
	assert.NotEmpty(t, got)
}

func TestCruelCuiControllerMoveInvalidDestination(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	got := c.Exec("m 0 xyz")
	assert.NotEmpty(t, got)
}

func TestCruelCuiControllerUnknownCommand(t *testing.T) {
	ci := newMockCruelInteractor()
	c := NewCruelCuiController(ci)
	got := c.Exec("zzz")
	assert.NotEmpty(t, got)
}

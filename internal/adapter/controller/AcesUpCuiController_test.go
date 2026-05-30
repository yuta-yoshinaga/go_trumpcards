package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockAcesUpInteractor() *mockusecase.MockAcesUpInteractor {
	return new(mockusecase.MockAcesUpInteractor)
}

func TestAcesUpCuiControllerQuit(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestAcesUpCuiControllerReset(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestAcesUpCuiControllerDraw(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestAcesUpCuiControllerGiveUp(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestAcesUpCuiControllerHint(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestAcesUpCuiControllerUndo(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestAcesUpCuiControllerActionLog(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestAcesUpCuiControllerRemove(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Remove", 2).Return("remove_output")
	assert.Equal(t, "remove_output", c.Exec("rm 2"))
	assert.Equal(t, "remove_output", c.Exec("remove 2"))
}

func TestAcesUpCuiControllerMove(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Move", 1).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("mv 1"))
	assert.Equal(t, "move_output", c.Exec("move 1"))
}

func TestAcesUpCuiControllerColCommand_InvalidArgs(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	assert.Contains(t, c.Exec("rm"), "Usage")
	assert.Contains(t, c.Exec("rm 3 0"), "Usage")
	assert.Contains(t, c.Exec("rm a"), "Invalid col")
	assert.Contains(t, c.Exec("mv"), "Usage")
	assert.Contains(t, c.Exec("mv x"), "Invalid col")
}

func TestAcesUpCuiControllerUnknown(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

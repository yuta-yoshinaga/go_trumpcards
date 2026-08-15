package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockPyramidInteractor() *mockusecase.MockPyramidInteractor {
	return new(mockusecase.MockPyramidInteractor)
}

func TestPyramidCuiControllerQuit(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestPyramidCuiControllerReset(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestPyramidCuiControllerDraw(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestPyramidCuiControllerGiveUp(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestPyramidCuiControllerHint(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestPyramidCuiControllerUndo(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestPyramidCuiControllerActionLog(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("l"))
	assert.Equal(t, "log_output", c.Exec("log"))
}

func TestPyramidCuiControllerRemoveKing(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("RemoveKing", 6, 0).Return("rk_output")
	assert.Equal(t, "rk_output", c.Exec("rm 6 0"))
}

func TestPyramidCuiControllerRemovePair(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("RemovePair", 6, 0, 6, 1).Return("rp_output")
	assert.Equal(t, "rp_output", c.Exec("rm 6 0 6 1"))
}

func TestPyramidCuiControllerRemoveWithWaste(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("RemoveWithWaste", 6, 0).Return("rw_output")
	assert.Equal(t, "rw_output", c.Exec("rm w 6 0"))
}

func TestPyramidCuiControllerRemoveWasteKing(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	pi.On("RemoveWasteKing").Return("rwk_output")
	assert.Equal(t, "rwk_output", c.Exec("rm w"))
}

func TestPyramidCuiControllerRemoveNoArgs(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	result := c.Exec("rm")
	assert.Contains(t, result, "Usage:")
}

func TestPyramidCuiControllerRemoveInvalidArgs(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	// Invalid row
	result := c.Exec("rm abc 0")
	assert.True(t, msgRejected(result))
	// Invalid col
	result = c.Exec("rm 6 abc")
	assert.True(t, msgRejected(result))
}

func TestPyramidCuiControllerRemovePairInvalidArgs(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	result := c.Exec("rm 6 0 abc 1")
	assert.True(t, msgRejected(result))
	result = c.Exec("rm 6 0 6 abc")
	assert.True(t, msgRejected(result))
}

func TestPyramidCuiControllerRemoveWasteInvalid(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	// rm w with 2 args (invalid)
	result := c.Exec("rm w 6")
	assert.Contains(t, result, "Usage:")
	// rm w with invalid number
	result = c.Exec("rm w abc 0")
	assert.True(t, msgRejected(result))
	result = c.Exec("rm w 6 abc")
	assert.True(t, msgRejected(result))
}

func TestPyramidCuiControllerRemove3Args(t *testing.T) {
	pi := newMockPyramidInteractor()
	c := NewPyramidCuiController(pi)
	result := c.Exec("rm 6 0 6")
	assert.Contains(t, result, "Usage:")
}

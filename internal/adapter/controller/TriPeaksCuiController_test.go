package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockTriPeaksInteractor() *mockusecase.MockTriPeaksInteractor {
	return new(mockusecase.MockTriPeaksInteractor)
}

func TestTriPeaksCuiControllerQuit(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestTriPeaksCuiControllerReset(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	ti.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestTriPeaksCuiControllerDraw(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	ti.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestTriPeaksCuiControllerGiveUp(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	ti.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestTriPeaksCuiControllerHint(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	ti.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestTriPeaksCuiControllerUndo(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	ti.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestTriPeaksCuiControllerActionLog(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	ti.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestTriPeaksCuiControllerRemove(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	ti.On("Remove", 3, 0).Return("remove_output")
	assert.Equal(t, "remove_output", c.Exec("rm 3 0"))
	assert.Equal(t, "remove_output", c.Exec("remove 3 0"))
}

func TestTriPeaksCuiControllerRemove_InvalidArgs(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	assert.Contains(t, c.Exec("rm"), "Usage")
	assert.Contains(t, c.Exec("rm 3"), "Usage")
	assert.True(t, msgRejected(c.Exec("rm a 0")))
	assert.True(t, msgRejected(c.Exec("rm 3 b")))
}

func TestTriPeaksCuiControllerUnknown(t *testing.T) {
	ti := newMockTriPeaksInteractor()
	c := NewTriPeaksCuiController(ti)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockGolfInteractor() *mockusecase.MockGolfInteractor {
	return new(mockusecase.MockGolfInteractor)
}

func TestGolfCuiControllerQuit(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestGolfCuiControllerReset(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	gi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestGolfCuiControllerDraw(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	gi.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestGolfCuiControllerGiveUp(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	gi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestGolfCuiControllerHint(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	gi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestGolfCuiControllerUndo(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	gi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestGolfCuiControllerActionLog(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	gi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestGolfCuiControllerRemove(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	gi.On("Remove", 3).Return("remove_output")
	assert.Equal(t, "remove_output", c.Exec("rm 3"))
	assert.Equal(t, "remove_output", c.Exec("remove 3"))
}

func TestGolfCuiControllerRemove_InvalidArgs(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	assert.True(t, msgRejected(c.Exec("rm")))
	assert.True(t, msgRejected(c.Exec("rm 3 0")))
	assert.True(t, msgRejected(c.Exec("rm a")))
}

func TestGolfCuiControllerUnknown(t *testing.T) {
	gi := newMockGolfInteractor()
	c := NewGolfCuiController(gi)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

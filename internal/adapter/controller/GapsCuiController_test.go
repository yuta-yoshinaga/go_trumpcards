package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockGapsInteractor() *mockusecase.MockGapsInteractor {
	return new(mockusecase.MockGapsInteractor)
}

func TestGapsCuiController_Quit(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestGapsCuiController_Reset(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	gi.On("Reset").Return("reset_out")
	assert.Equal(t, "reset_out", c.Exec("r"))
	assert.Equal(t, "reset_out", c.Exec("reset"))
}

func TestGapsCuiController_Move(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	gi.On("Move", 0, 1, 2, 3).Return("move_out")
	assert.Equal(t, "move_out", c.Exec("m 0 1 2 3"))
	assert.Equal(t, "move_out", c.Exec("move 0 1 2 3"))
}

func TestGapsCuiController_Move_BadArgs(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	assert.Contains(t, c.Exec("m"), msgUsage("usageMFromrowFromcolTorowTocol"))
	assert.Contains(t, c.Exec("m 0 1 2"), msgUsage("usageMFromrowFromcolTorowTocol"))
	assert.True(t, msgRejected(c.Exec("m x 1 2 3")))
	assert.True(t, msgRejected(c.Exec("m 0 x 2 3")))
	assert.True(t, msgRejected(c.Exec("m 0 1 x 3")))
	assert.True(t, msgRejected(c.Exec("m 0 1 2 x")))
}

func TestGapsCuiController_Redeal(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	gi.On("Redeal").Return("rd_out")
	assert.Equal(t, "rd_out", c.Exec("rd"))
	assert.Equal(t, "rd_out", c.Exec("redeal"))
}

func TestGapsCuiController_GiveUp(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	gi.On("GiveUp").Return("giveup_out")
	assert.Equal(t, "giveup_out", c.Exec("g"))
	assert.Equal(t, "giveup_out", c.Exec("giveup"))
}

func TestGapsCuiController_Undo(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	gi.On("Undo").Return("undo_out")
	assert.Equal(t, "undo_out", c.Exec("u"))
	assert.Equal(t, "undo_out", c.Exec("undo"))
}

func TestGapsCuiController_Hint(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	gi.On("Hint").Return("hint_out")
	assert.Equal(t, "hint_out", c.Exec("h"))
	assert.Equal(t, "hint_out", c.Exec("hint"))
}

func TestGapsCuiController_ActionLog(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	gi.On("ActionLog").Return("log_out")
	assert.Equal(t, "log_out", c.Exec("l"))
	assert.Equal(t, "log_out", c.Exec("log"))
}

func TestGapsCuiController_Unknown(t *testing.T) {
	gi := newMockGapsInteractor()
	c := NewGapsCuiController(gi)
	assert.NotEmpty(t, c.Exec("xxx"))
}

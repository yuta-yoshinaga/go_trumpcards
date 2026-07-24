package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockClockSolitaireInteractor() *mockusecase.MockClockSolitaireInteractor {
	return new(mockusecase.MockClockSolitaireInteractor)
}

func TestClockSolitaireCuiControllerQuit(t *testing.T) {
	ci := newMockClockSolitaireInteractor()
	c := NewClockSolitaireCuiController(ci)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestClockSolitaireCuiControllerReset(t *testing.T) {
	ci := newMockClockSolitaireInteractor()
	c := NewClockSolitaireCuiController(ci)
	ci.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestClockSolitaireCuiControllerStep(t *testing.T) {
	ci := newMockClockSolitaireInteractor()
	c := NewClockSolitaireCuiController(ci)
	ci.On("Step").Return("step_output")
	assert.Equal(t, "step_output", c.Exec("s"))
	assert.Equal(t, "step_output", c.Exec("step"))
}

func TestClockSolitaireCuiControllerAutoPlay(t *testing.T) {
	ci := newMockClockSolitaireInteractor()
	c := NewClockSolitaireCuiController(ci)
	ci.On("AutoPlay").Return("autoplay_output")
	assert.Equal(t, "autoplay_output", c.Exec("a"))
	assert.Equal(t, "autoplay_output", c.Exec("autoplay"))
}

func TestClockSolitaireCuiControllerUndo(t *testing.T) {
	ci := newMockClockSolitaireInteractor()
	c := NewClockSolitaireCuiController(ci)
	ci.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestClockSolitaireCuiControllerActionLog(t *testing.T) {
	ci := newMockClockSolitaireInteractor()
	c := NewClockSolitaireCuiController(ci)
	ci.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestClockSolitaireCuiControllerEmpty(t *testing.T) {
	ci := newMockClockSolitaireInteractor()
	c := NewClockSolitaireCuiController(ci)
	result := c.Exec("")
	assert.NotEmpty(t, result)
}

func TestClockSolitaireCuiControllerUnknown(t *testing.T) {
	ci := newMockClockSolitaireInteractor()
	c := NewClockSolitaireCuiController(ci)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

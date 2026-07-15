package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockTrashInteractor() *mockusecase.MockTrashInteractor {
	return new(mockusecase.MockTrashInteractor)
}

func TestTrashCuiControllerQuit(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestTrashCuiControllerReset(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	ti.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestTrashCuiControllerDraw(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	ti.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestTrashCuiControllerPlace(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	ti.On("PlaceWild", 3).Return("place_output")
	assert.Equal(t, "place_output", c.Exec("p 3"))
	assert.Equal(t, "place_output", c.Exec("place 3"))
}

func TestTrashCuiControllerPlacePrompt(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	assert.Contains(t, c.Exec("p"), cuiutil.PromptPrefix)
}

func TestTrashCuiControllerPlaceInvalid(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	assert.NotEmpty(t, c.Exec("p abc"))
}

func TestTrashCuiControllerCpu(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	ti.On("CpuStep").Return("cpu_output")
	assert.Equal(t, "cpu_output", c.Exec("cpu"))
}

func TestTrashCuiControllerHint(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	ti.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestTrashCuiControllerActionLog(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	ti.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestTrashCuiControllerUnknown(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	assert.NotEmpty(t, c.Exec("unknowncmd"))
}

func TestTrashCuiControllerEmpty(t *testing.T) {
	ti := newMockTrashInteractor()
	c := NewTrashCuiController(ti)
	assert.NotEmpty(t, c.Exec(""))
}

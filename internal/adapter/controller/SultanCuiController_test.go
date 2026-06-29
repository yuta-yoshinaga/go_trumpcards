package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockSultanInteractor() *mockusecase.MockSultanInteractor {
	return new(mockusecase.MockSultanInteractor)
}

func TestSultanCuiControllerQuit(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSultanCuiControllerReset(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestSultanCuiControllerDraw(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestSultanCuiControllerRedeal(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("Redeal").Return("redeal_output")
	assert.Equal(t, "redeal_output", c.Exec("rd"))
	assert.Equal(t, "redeal_output", c.Exec("redeal"))
}

func TestSultanCuiControllerGiveUp(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestSultanCuiControllerHint(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestSultanCuiControllerAutoComplete(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestSultanCuiControllerActionLog(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestSultanCuiControllerUndo(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestSultanCuiControllerMoveDivanToFoundation(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("MoveDivanToFoundation", 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m d 3"))
}

func TestSultanCuiControllerMoveWasteToFoundation(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("MoveWasteToFoundation").Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m w"))
}

func TestSultanCuiControllerMoveShorthand(t *testing.T) {
	si := newMockSultanInteractor()
	c := NewSultanCuiController(si)
	si.On("MoveDivanToFoundation", 2).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 2"))
}

func TestSultanCuiControllerMoveErrors(t *testing.T) {
	t.Run("move no args prompts", func(t *testing.T) {
		si := newMockSultanInteractor()
		c := NewSultanCuiController(si)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m")))
	})

	t.Run("move divan no index prompts", func(t *testing.T) {
		si := newMockSultanInteractor()
		c := NewSultanCuiController(si)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m d")))
	})

	t.Run("move divan invalid index", func(t *testing.T) {
		si := newMockSultanInteractor()
		c := NewSultanCuiController(si)
		result := c.Exec("m d abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("move invalid from zone", func(t *testing.T) {
		si := newMockSultanInteractor()
		c := NewSultanCuiController(si)
		result := c.Exec("m x")
		assert.Contains(t, result, "x")
	})
}

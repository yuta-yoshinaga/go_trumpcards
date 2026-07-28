package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockBisleyInteractor() *mockusecase.MockBisleyInteractor {
	return new(mockusecase.MockBisleyInteractor)
}

func TestBisleyCuiControllerQuit(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestBisleyCuiControllerReset(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestBisleyCuiControllerGiveUp(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestBisleyCuiControllerHint(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestBisleyCuiControllerAutoComplete(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestBisleyCuiControllerActionLog(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestBisleyCuiControllerUndo(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestBisleyCuiControllerMoveTableauToAceFoundation(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("MoveTableauToAceFoundation", 2).Return("ace_output")
	assert.Equal(t, "ace_output", c.Exec("m 2 a"))
}

func TestBisleyCuiControllerMoveTableauToKingFoundation(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("MoveTableauToKingFoundation", 4).Return("king_output")
	assert.Equal(t, "king_output", c.Exec("m 4 k"))
}

func TestBisleyCuiControllerMoveTableauToTableau(t *testing.T) {
	bi := newMockBisleyInteractor()
	c := NewBisleyCuiController(bi)
	bi.On("MoveTableauToTableau", 0, 5).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 5"))
}

func TestBisleyCuiControllerMoveErrors(t *testing.T) {
	t.Run("move no args prompts", func(t *testing.T) {
		bi := newMockBisleyInteractor()
		c := NewBisleyCuiController(bi)
		result := c.Exec("m")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid from col", func(t *testing.T) {
		bi := newMockBisleyInteractor()
		c := NewBisleyCuiController(bi)
		result := c.Exec("m abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("move single arg prompts for destination", func(t *testing.T) {
		bi := newMockBisleyInteractor()
		c := NewBisleyCuiController(bi)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid to col", func(t *testing.T) {
		bi := newMockBisleyInteractor()
		c := NewBisleyCuiController(bi)
		result := c.Exec("m 0 xyz")
		assert.Contains(t, result, "xyz")
	})
}

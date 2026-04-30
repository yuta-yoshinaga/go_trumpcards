package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockBakersDozenInteractor() *mockusecase.MockBakersDozenInteractor {
	return new(mockusecase.MockBakersDozenInteractor)
}

func TestBakersDozenCuiControllerQuit(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestBakersDozenCuiControllerReset(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	bi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestBakersDozenCuiControllerGiveUp(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	bi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestBakersDozenCuiControllerHint(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	bi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestBakersDozenCuiControllerAutoComplete(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	bi.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestBakersDozenCuiControllerActionLog(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	bi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestBakersDozenCuiControllerUndo(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	bi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestBakersDozenCuiControllerMoveTableauToFoundation(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	bi.On("MoveTableauToFoundation", 2).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 2 f"))
}

func TestBakersDozenCuiControllerMoveTableauToTableau(t *testing.T) {
	bi := newMockBakersDozenInteractor()
	c := NewBakersDozenCuiController(bi)
	bi.On("MoveTableauToTableau", 0, -1, 5).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 5"))
}

func TestBakersDozenCuiControllerMoveErrors(t *testing.T) {
	t.Run("move no args prompts", func(t *testing.T) {
		bi := newMockBakersDozenInteractor()
		c := NewBakersDozenCuiController(bi)
		result := c.Exec("m")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid from col", func(t *testing.T) {
		bi := newMockBakersDozenInteractor()
		c := NewBakersDozenCuiController(bi)
		result := c.Exec("m abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("move single arg prompts for destination", func(t *testing.T) {
		bi := newMockBakersDozenInteractor()
		c := NewBakersDozenCuiController(bi)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid to col", func(t *testing.T) {
		bi := newMockBakersDozenInteractor()
		c := NewBakersDozenCuiController(bi)
		result := c.Exec("m 0 xyz")
		assert.Contains(t, result, "xyz")
	})
}

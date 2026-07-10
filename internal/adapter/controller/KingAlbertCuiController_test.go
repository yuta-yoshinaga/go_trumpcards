package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockKingAlbertInteractor() *mockusecase.MockKingAlbertInteractor {
	return new(mockusecase.MockKingAlbertInteractor)
}

func TestKingAlbertCuiControllerQuit(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestKingAlbertCuiControllerReset(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestKingAlbertCuiControllerGiveUp(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestKingAlbertCuiControllerHint(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestKingAlbertCuiControllerAutoComplete(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestKingAlbertCuiControllerActionLog(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestKingAlbertCuiControllerUndo(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestKingAlbertCuiControllerMoveTableauToFoundation(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("MoveTableauToFoundation", 2).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 2 f"))
}

func TestKingAlbertCuiControllerMoveTableauToTableau(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("MoveTableauToTableau", 0, -1, 5).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 5"))
}

func TestKingAlbertCuiControllerMoveReserveToTableau(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("MoveReserveToTableau", 1, 4).Return("rt_output")
	assert.Equal(t, "rt_output", c.Exec("m r1 4"))
}

func TestKingAlbertCuiControllerMoveReserveToFoundation(t *testing.T) {
	bi := newMockKingAlbertInteractor()
	c := NewKingAlbertCuiController(bi)
	bi.On("MoveReserveToFoundation", 0).Return("rf_output")
	assert.Equal(t, "rf_output", c.Exec("m r0 f"))
}

func TestKingAlbertCuiControllerMoveErrors(t *testing.T) {
	t.Run("move no args prompts", func(t *testing.T) {
		bi := newMockKingAlbertInteractor()
		c := NewKingAlbertCuiController(bi)
		result := c.Exec("m")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid from col", func(t *testing.T) {
		bi := newMockKingAlbertInteractor()
		c := NewKingAlbertCuiController(bi)
		result := c.Exec("m abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("move single arg prompts for destination", func(t *testing.T) {
		bi := newMockKingAlbertInteractor()
		c := NewKingAlbertCuiController(bi)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid to col", func(t *testing.T) {
		bi := newMockKingAlbertInteractor()
		c := NewKingAlbertCuiController(bi)
		result := c.Exec("m 0 xyz")
		assert.Contains(t, result, "xyz")
	})

	t.Run("reserve invalid index", func(t *testing.T) {
		bi := newMockKingAlbertInteractor()
		c := NewKingAlbertCuiController(bi)
		result := c.Exec("m rx 4")
		assert.Contains(t, result, "rx")
	})

	t.Run("reserve single arg prompts", func(t *testing.T) {
		bi := newMockKingAlbertInteractor()
		c := NewKingAlbertCuiController(bi)
		result := c.Exec("m r0")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("reserve invalid to col", func(t *testing.T) {
		bi := newMockKingAlbertInteractor()
		c := NewKingAlbertCuiController(bi)
		result := c.Exec("m r0 xyz")
		assert.Contains(t, result, "xyz")
	})
}

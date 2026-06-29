package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockFortyAndEightInteractor() *mockusecase.MockFortyAndEightInteractor {
	return new(mockusecase.MockFortyAndEightInteractor)
}

func TestFortyAndEightCuiControllerQuit(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestFortyAndEightCuiControllerReset(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestFortyAndEightCuiControllerDraw(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestFortyAndEightCuiControllerRedeal(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("Redeal").Return("redeal_output")
	assert.Equal(t, "redeal_output", c.Exec("rd"))
	assert.Equal(t, "redeal_output", c.Exec("redeal"))
}

func TestFortyAndEightCuiControllerGiveUp(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestFortyAndEightCuiControllerHint(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestFortyAndEightCuiControllerAutoComplete(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestFortyAndEightCuiControllerActionLog(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestFortyAndEightCuiControllerUndo(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestFortyAndEightCuiControllerMoveWasteToTableau(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("MoveWasteToTableau", 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m w t 3"))
}

func TestFortyAndEightCuiControllerMoveWasteToFoundation(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("MoveWasteToFoundation").Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m w f"))
}

func TestFortyAndEightCuiControllerMoveTableauToFoundation(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("MoveTableauToFoundation", 2).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 2 f"))
}

func TestFortyAndEightCuiControllerMoveTableauToTableau(t *testing.T) {
	fi := newMockFortyAndEightInteractor()
	c := NewFortyAndEightCuiController(fi)
	fi.On("MoveTableauToTableau", 0, 3, 5).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 0 3 t 5"))
}

func TestFortyAndEightCuiControllerMoveErrors(t *testing.T) {
	t.Run("move too few args", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move one arg w", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m w")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid from", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m x t 3")
		assert.Contains(t, result, "x")
	})

	t.Run("move waste invalid to", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m w x")
		assert.Contains(t, result, "x")
	})

	t.Run("move waste to tableau no col", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m w t")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move waste to tableau invalid col", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m w t abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("move tableau too few args", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m t")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move tableau one arg", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m t 5")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move tableau invalid from col", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m t abc f")
		assert.Contains(t, result, "abc")
	})

	t.Run("move tableau bad format", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m t 0 3 x 5")
		assert.NotEmpty(t, result)
	})

	t.Run("move tableau wizard prompt toCol", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m t 0 3 t")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move tableau invalid cardIdx", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m t 0 abc t 5")
		assert.Contains(t, result, "abc")
	})

	t.Run("move tableau invalid toCol", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m t 0 3 t xyz")
		assert.Contains(t, result, "xyz")
	})
}

func TestFortyAndEightCuiControllerMoveShorthand(t *testing.T) {
	t.Run("m <from> <to> moves top card", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		fi.On("MoveTableauToTableau", 0, -1, 1).Return("move_output")
		assert.Equal(t, "move_output", c.Exec("m 0 1"))
	})

	t.Run("m <from> prompts for destination", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "m 0 {0}", tmpl)
	})

	t.Run("m <from> <invalid> returns error", func(t *testing.T) {
		fi := newMockFortyAndEightInteractor()
		c := NewFortyAndEightCuiController(fi)
		result := c.Exec("m 0 abc")
		assert.Contains(t, result, "abc")
	})
}

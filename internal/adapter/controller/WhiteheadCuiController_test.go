package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockWhiteheadInteractor() *mockusecase.MockWhiteheadInteractor {
	return new(mockusecase.MockWhiteheadInteractor)
}

func TestWhiteheadCuiControllerQuit(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestWhiteheadCuiControllerReset(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestWhiteheadCuiControllerDraw(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestWhiteheadCuiControllerGiveUp(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestWhiteheadCuiControllerHint(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestWhiteheadCuiControllerAutoComplete(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestWhiteheadCuiControllerActionLog(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestWhiteheadCuiControllerMoveWasteToTableau(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("MoveWasteToTableau", 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m w t 3"))
}

func TestWhiteheadCuiControllerMoveWasteToFoundation(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("MoveWasteToFoundation").Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m w f"))
}

func TestWhiteheadCuiControllerMoveTableauToFoundation(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("MoveTableauToFoundation", 2).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 2 f"))
}

func TestWhiteheadCuiControllerMoveTableauToTableau(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("MoveTableauToTableau", 0, 3, 4).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 0 3 t 4"))
}

func TestWhiteheadCuiControllerMoveErrors(t *testing.T) {
	t.Run("move too few args", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move one arg", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m w")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid from", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m x t 3")
		assert.Contains(t, result, "x")
	})

	t.Run("move waste invalid to", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m w x")
		assert.Contains(t, result, "x")
	})

	t.Run("move waste to tableau no col", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m w t")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move waste to tableau invalid col", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m w t abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("move tableau too few args", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m t")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move tableau one arg - too few for handleMoveFromTableau", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m t 5")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move tableau invalid from col", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m t abc f")
		assert.Contains(t, result, "abc")
	})

	t.Run("move tableau invalid second arg (not f and not enough args)", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m t 0 x")
		assert.NotEmpty(t, result)
		assert.False(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move tableau to tableau chained wizard toCol prompt", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m t 0 3 t")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "m t 0 3 t {0}", tmpl)
	})

	t.Run("move tableau to tableau wrong zone marker", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m t 0 3 f 4")
		assert.NotEmpty(t, result)
	})

	t.Run("move tableau to tableau invalid card index", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m t 0 abc t 4")
		assert.Contains(t, result, "abc")
	})

	t.Run("move tableau to tableau invalid to col", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m t 0 3 t abc")
		assert.Contains(t, result, "abc")
	})
}

func TestWhiteheadCuiControllerMoveShorthand(t *testing.T) {
	t.Run("m <from> <to> moves top card", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		ki.On("MoveTableauToTableau", 0, -1, 1).Return("move_output")
		assert.Equal(t, "move_output", c.Exec("m 0 1"))
	})

	t.Run("m <from> prompts for destination", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "m 0 {0}", tmpl)
	})

	t.Run("m <from> <invalid> returns error", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("m 0 abc")
		assert.Contains(t, result, "abc")
	})
}

func TestWhiteheadCuiControllerFoundationShorthand(t *testing.T) {
	t.Run("f with no args moves waste to foundation", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		ki.On("MoveWasteToFoundation").Return("wf_output")
		assert.Equal(t, "wf_output", c.Exec("f"))
	})

	t.Run("f <col> moves tableau to foundation", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		ki.On("MoveTableauToFoundation", 3).Return("tf_output")
		assert.Equal(t, "tf_output", c.Exec("f 3"))
	})

	t.Run("f <invalid> returns error", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		result := c.Exec("f abc")
		assert.Contains(t, result, "abc")
	})
}

func TestWhiteheadCuiControllerUnknown(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestWhiteheadCuiControllerEmpty(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

func TestWhiteheadCuiControllerResetWithConfig(t *testing.T) {
	t.Run("reset 3", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		ki.On("ResetWithConfig", domain.WhiteheadConfig{DrawCount: 3}).Return("reset3_output")
		assert.Equal(t, "reset3_output", c.Exec("reset 3"))
	})

	t.Run("reset 1", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		ki.On("ResetWithConfig", domain.WhiteheadConfig{DrawCount: 1}).Return("reset1_output")
		assert.Equal(t, "reset1_output", c.Exec("reset 1"))
	})

	t.Run("reset abc falls back to Reset", func(t *testing.T) {
		ki := newMockWhiteheadInteractor()
		c := NewWhiteheadCuiController(ki)
		ki.On("Reset").Return("reset_output")
		assert.Equal(t, "reset_output", c.Exec("reset abc"))
	})
}

func TestWhiteheadCuiControllerUndo(t *testing.T) {
	ki := newMockWhiteheadInteractor()
	c := NewWhiteheadCuiController(ki)
	ki.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

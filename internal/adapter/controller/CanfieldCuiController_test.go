package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCanfieldInteractor() *mockusecase.MockCanfieldInteractor {
	return new(mockusecase.MockCanfieldInteractor)
}

func TestCanfieldCuiControllerQuit(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCanfieldCuiControllerReset(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("Reset").Return("reset")
	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "reset", c.Exec("reset"))
}

func TestCanfieldCuiControllerDraw(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("Draw").Return("draw")
	assert.Equal(t, "draw", c.Exec("d"))
	assert.Equal(t, "draw", c.Exec("draw"))
}

func TestCanfieldCuiControllerGiveUp(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("GiveUp").Return("giveup")
	assert.Equal(t, "giveup", c.Exec("g"))
	assert.Equal(t, "giveup", c.Exec("giveup"))
}

func TestCanfieldCuiControllerHint(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("Hint").Return("hint")
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "hint", c.Exec("hint"))
}

func TestCanfieldCuiControllerAutoComplete(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("AutoComplete").Return("ac")
	assert.Equal(t, "ac", c.Exec("ac"))
	assert.Equal(t, "ac", c.Exec("autocomplete"))
}

func TestCanfieldCuiControllerLog(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("ActionLog").Return("log")
	assert.Equal(t, "log", c.Exec("log"))
	assert.Equal(t, "log", c.Exec("l"))
}

func TestCanfieldCuiControllerUndo(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("Undo").Return("undo")
	assert.Equal(t, "undo", c.Exec("u"))
	assert.Equal(t, "undo", c.Exec("undo"))
}

func TestCanfieldCuiControllerMoveWasteToTableau(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("MoveWasteToTableau", 2).Return("mwt")
	assert.Equal(t, "mwt", c.Exec("m w t 2"))
}

func TestCanfieldCuiControllerMoveWasteToFoundation(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("MoveWasteToFoundation").Return("mwf")
	assert.Equal(t, "mwf", c.Exec("m w f"))
}

func TestCanfieldCuiControllerMoveReserveToTableau(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("MoveReserveToTableau", 1).Return("mrt")
	assert.Equal(t, "mrt", c.Exec("m r t 1"))
}

func TestCanfieldCuiControllerMoveReserveToFoundation(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("MoveReserveToFoundation").Return("mrf")
	assert.Equal(t, "mrf", c.Exec("m r f"))
}

func TestCanfieldCuiControllerMoveTableauToFoundation(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("MoveTableauToFoundation", 2).Return("mtf")
	assert.Equal(t, "mtf", c.Exec("m t 2 f"))
}

func TestCanfieldCuiControllerMoveTableauToTableau(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	ci.On("MoveTableauToTableau", 0, 1, 2).Return("mtt")
	assert.Equal(t, "mtt", c.Exec("m t 0 1 t 2"))
}

func TestCanfieldCuiControllerMoveErrors(t *testing.T) {
	t.Run("empty move", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m")))
	})
	t.Run("move w only", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m w")))
	})
	t.Run("move r only", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m r")))
	})
	t.Run("move t only", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t")))
	})
	t.Run("invalid from", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.Contains(t, c.Exec("m x t 0"), "x")
	})
	t.Run("move w t no col", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m w t")))
	})
	t.Run("move w invalid to", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.Contains(t, c.Exec("m w x"), "x")
	})
	t.Run("move r invalid to", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.Contains(t, c.Exec("m r x"), "x")
	})
	t.Run("move r t no col", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m r t")))
	})
	t.Run("move w t invalid col", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.Contains(t, c.Exec("m w t abc"), "abc")
	})
	t.Run("move r t invalid col", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.Contains(t, c.Exec("m r t abc"), "abc")
	})
	t.Run("move t invalid fromcol", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.Contains(t, c.Exec("m t abc f"), "abc")
	})
	t.Run("move t wizard to tableau", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t 0 1 t")))
	})
	t.Run("move t invalid idx", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.Contains(t, c.Exec("m t 0 abc t 1"), "abc")
	})
	t.Run("move t invalid to col", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.Contains(t, c.Exec("m t 0 1 t abc"), "abc")
	})
	t.Run("move t wrong marker", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.NotEmpty(t, c.Exec("m t 0 1 x 2"))
	})
	t.Run("move t one arg", func(t *testing.T) {
		ci := newMockCanfieldInteractor()
		c := NewCanfieldCuiController(ci)
		assert.True(t, cuiutil.IsPromptRequest(c.Exec("m t 0")))
	})
}

func TestCanfieldCuiControllerUnknown(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

func TestCanfieldCuiControllerEmpty(t *testing.T) {
	ci := newMockCanfieldInteractor()
	c := NewCanfieldCuiController(ci)
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

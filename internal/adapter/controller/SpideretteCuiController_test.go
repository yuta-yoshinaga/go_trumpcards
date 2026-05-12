package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockSpideretteInteractor() *mockusecase.MockSpideretteInteractor {
	return new(mockusecase.MockSpideretteInteractor)
}

func TestSpideretteCuiControllerQuit(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSpideretteCuiControllerReset(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	si.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestSpideretteCuiControllerDeal(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	si.On("Deal").Return("deal_output")
	assert.Equal(t, "deal_output", c.Exec("d"))
	assert.Equal(t, "deal_output", c.Exec("deal"))
}

func TestSpideretteCuiControllerMoveTableauToTableau(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	si.On("MoveTableauToTableau", 0, 3, 4).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 0 3 t 4"))
}

func TestSpideretteCuiControllerMoveErrors(t *testing.T) {
	t.Run("move too few args prompts", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move with bogus marker", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m x 0 3 t 4")
		assert.NotEmpty(t, result)
		assert.False(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid from col", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m t abc 3 t 4")
		assert.Contains(t, result, "abc")
	})

	t.Run("move wrong arg count prompts", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m t 0 3")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move missing t marker", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m t 0 3 f 4")
		assert.NotEmpty(t, result)
		assert.False(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid card index", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m t 0 abc t 4")
		assert.Contains(t, result, "abc")
	})

	t.Run("move invalid to col", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m t 0 3 t abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("move chained wizard 4-arg toCol prompt", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m t 0 3 t")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Contains(t, tmpl, "m t 0 3 t {0}")
	})
}

func TestSpideretteCuiControllerGiveUp(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	si.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestSpideretteCuiControllerHint(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	si.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestSpideretteCuiControllerAutoComplete(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	si.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestSpideretteCuiControllerActionLog(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	si.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestSpideretteCuiControllerUndo(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	si.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestSpideretteCuiControllerMoveShorthand(t *testing.T) {
	t.Run("m <from> <to> moves top card", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		si.On("MoveTableauToTableau", 0, -1, 1).Return("move_output")
		assert.Equal(t, "move_output", c.Exec("m 0 1"))
	})

	t.Run("m <from> <idx> <to>", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		si.On("MoveTableauToTableau", 0, 2, 1).Return("move_output")
		assert.Equal(t, "move_output", c.Exec("m 0 2 1"))
	})

	t.Run("m <from> prompts for destination", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "m 0 {0}", tmpl)
	})

	t.Run("m <from> <invalid> returns error", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m 0 abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("m <from> <idx> <invalid> returns error", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m 0 2 abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("m <from> <invalidIdx> <to> returns error", func(t *testing.T) {
		si := newMockSpideretteInteractor()
		c := NewSpideretteCuiController(si)
		result := c.Exec("m 0 abc 1")
		assert.Contains(t, result, "abc")
	})
}

func TestSpideretteCuiControllerUnknown(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestSpideretteCuiControllerEmpty(t *testing.T) {
	si := newMockSpideretteInteractor()
	c := NewSpideretteCuiController(si)
	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

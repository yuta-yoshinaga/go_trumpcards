package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockMrsMopInteractor() *mockusecase.MockMrsMopInteractor {
	return new(mockusecase.MockMrsMopInteractor)
}

func TestMrsMopCuiControllerQuit(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestMrsMopCuiControllerReset(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	si.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestMrsMopCuiControllerResetWithConfig(t *testing.T) {
	t.Run("reset 1", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		si.On("ResetWithConfig", domain.MrsMopConfig{Difficulty: domain.MrsMopDifficulty1Suit}).Return("reset1_output")
		assert.Equal(t, "reset1_output", c.Exec("reset 1"))
	})

	t.Run("reset 2", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		si.On("ResetWithConfig", domain.MrsMopConfig{Difficulty: domain.MrsMopDifficulty2Suit}).Return("reset2_output")
		assert.Equal(t, "reset2_output", c.Exec("reset 2"))
	})

	t.Run("reset 4", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		si.On("ResetWithConfig", domain.MrsMopConfig{Difficulty: domain.MrsMopDifficulty4Suit}).Return("reset4_output")
		assert.Equal(t, "reset4_output", c.Exec("reset 4"))
	})

	t.Run("reset abc falls back to Reset", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		si.On("Reset").Return("reset_output")
		assert.Equal(t, "reset_output", c.Exec("reset abc"))
	})

	t.Run("reset 3 falls back to Reset (invalid difficulty)", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		si.On("Reset").Return("reset_output")
		assert.Equal(t, "reset_output", c.Exec("reset 3"))
	})
}

// **配るコマンドは存在しない。**山札が無いので、受け付けると盤が変わらない
// 無言の no-op になる。クローン元の Spider は 5 回配れる。
func TestMrsMopCuiControllerDealIsNotACommand(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	assert.Contains(t, c.Exec("d"), "コマンドが不明です")
	assert.Contains(t, c.Exec("deal"), "コマンドが不明です")
}

func TestMrsMopCuiControllerMoveTableauToTableau(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	si.On("MoveTableauToTableau", 0, 3, 4).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 0 3 t 4"))
}

func TestMrsMopCuiControllerMoveErrors(t *testing.T) {
	t.Run("move too few args", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move one arg not t", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m x 0 3 t 4")
		assert.NotEmpty(t, result)
		assert.False(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid from col", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m t abc 3 t 4")
		assert.Contains(t, result, "abc")
	})

	t.Run("move wrong arg count", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m t 0 3")
		assert.True(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move missing t marker", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m t 0 3 f 4")
		assert.NotEmpty(t, result)
		assert.False(t, cuiutil.IsPromptRequest(result))
	})

	t.Run("move invalid card index", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m t 0 abc t 4")
		assert.Contains(t, result, "abc")
	})

	t.Run("move invalid to col", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m t 0 3 t abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("move chained wizard 4-arg toCol prompt", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m t 0 3 t")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Contains(t, tmpl, "m t 0 3 t {0}")
	})
}

func TestMrsMopCuiControllerGiveUp(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	si.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestMrsMopCuiControllerHint(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	si.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestMrsMopCuiControllerAutoComplete(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	si.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestMrsMopCuiControllerActionLog(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	si.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestMrsMopCuiControllerUndo(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	si.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestMrsMopCuiControllerMoveShorthand(t *testing.T) {
	t.Run("m <from> <to> moves top card", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		si.On("MoveTableauToTableau", 0, -1, 1).Return("move_output")
		assert.Equal(t, "move_output", c.Exec("m 0 1"))
	})

	t.Run("m <from> <idx> <to> moves with card index", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		si.On("MoveTableauToTableau", 0, 2, 1).Return("move_output")
		assert.Equal(t, "move_output", c.Exec("m 0 2 1"))
	})

	t.Run("m <from> prompts for destination", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "m 0 {0}", tmpl)
	})

	t.Run("m <from> <invalid> returns error", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m 0 abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("m <from> <idx> <invalid> returns error", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m 0 2 abc")
		assert.Contains(t, result, "abc")
	})

	t.Run("m <from> <invalidIdx> <to> returns error", func(t *testing.T) {
		si := newMockMrsMopInteractor()
		c := NewMrsMopCuiController(si)
		result := c.Exec("m 0 abc 1")
		assert.Contains(t, result, "abc")
	})
}

func TestMrsMopCuiControllerUnknown(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestMrsMopCuiControllerEmpty(t *testing.T) {
	si := newMockMrsMopInteractor()
	c := NewMrsMopCuiController(si)
	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

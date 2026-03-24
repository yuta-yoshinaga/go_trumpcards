package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newMockKlondikeInteractor() *mockusecase.MockKlondikeInteractor {
	return new(mockusecase.MockKlondikeInteractor)
}

func TestKlondikeCuiControllerQuit(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestKlondikeCuiControllerReset(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestKlondikeCuiControllerDraw(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestKlondikeCuiControllerGiveUp(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestKlondikeCuiControllerHint(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestKlondikeCuiControllerAutoComplete(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestKlondikeCuiControllerActionLog(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestKlondikeCuiControllerMoveWasteToTableau(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("MoveWasteToTableau", 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m w t 3"))
}

func TestKlondikeCuiControllerMoveWasteToFoundation(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("MoveWasteToFoundation").Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m w f"))
}

func TestKlondikeCuiControllerMoveTableauToFoundation(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("MoveTableauToFoundation", 2).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 2 f"))
}

func TestKlondikeCuiControllerMoveTableauToTableau(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("MoveTableauToTableau", 0, 3, 4).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m t 0 3 t 4"))
}

func TestKlondikeCuiControllerMoveErrors(t *testing.T) {
	t.Run("move too few args", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m")
		assert.Contains(t, result, "Usage")
	})

	t.Run("move one arg", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m w")
		assert.Contains(t, result, "Usage")
	})

	t.Run("move invalid from", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m x t 3")
		assert.Contains(t, result, "Invalid from zone")
	})

	t.Run("move waste invalid to", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m w x")
		assert.Contains(t, result, "Invalid to zone")
	})

	t.Run("move waste to tableau no col", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m w t")
		assert.Contains(t, result, "Usage")
	})

	t.Run("move waste to tableau invalid col", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m w t abc")
		assert.Contains(t, result, "Invalid column")
	})

	t.Run("move tableau too few args", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m t")
		assert.Contains(t, result, "Usage")
	})

	t.Run("move tableau one arg - too few for handleMoveFromTableau", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m t 5")
		assert.Contains(t, result, "Usage")
	})

	t.Run("move tableau invalid from col", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m t abc f")
		assert.Contains(t, result, "Invalid from column")
	})

	t.Run("move tableau invalid second arg (not f and not enough args)", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m t 0 x")
		assert.Contains(t, result, "Invalid move command")
	})

	t.Run("move tableau to tableau too few args", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m t 0 3 t")
		assert.Contains(t, result, "Invalid move command")
	})

	t.Run("move tableau to tableau wrong zone marker", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m t 0 3 f 4")
		assert.Contains(t, result, "Invalid move command")
	})

	t.Run("move tableau to tableau invalid card index", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m t 0 abc t 4")
		assert.Contains(t, result, "Invalid card index")
	})

	t.Run("move tableau to tableau invalid to col", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		result := c.Exec("m t 0 3 t abc")
		assert.Contains(t, result, "Invalid to column")
	})
}

func TestKlondikeCuiControllerUnknown(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	result := c.Exec("xyz")
	assert.Contains(t, result, "コマンドが不明です")
}

func TestKlondikeCuiControllerEmpty(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	result := c.Exec("")
	assert.Contains(t, result, "'help' でコマンド一覧を表示します。")
}

func TestKlondikeCuiControllerResetWithConfig(t *testing.T) {
	t.Run("reset 3", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		ki.On("ResetWithConfig", domain.KlondikeConfig{DrawCount: 3}).Return("reset3_output")
		assert.Equal(t, "reset3_output", c.Exec("reset 3"))
	})

	t.Run("reset 1", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		ki.On("ResetWithConfig", domain.KlondikeConfig{DrawCount: 1}).Return("reset1_output")
		assert.Equal(t, "reset1_output", c.Exec("reset 1"))
	})

	t.Run("reset abc falls back to Reset", func(t *testing.T) {
		ki := newMockKlondikeInteractor()
		c := NewKlondikeCuiController(ki)
		ki.On("Reset").Return("reset_output")
		assert.Equal(t, "reset_output", c.Exec("reset abc"))
	})
}

func TestKlondikeCuiControllerUndo(t *testing.T) {
	ki := newMockKlondikeInteractor()
	c := NewKlondikeCuiController(ki)
	ki.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

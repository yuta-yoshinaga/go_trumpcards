//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockPenguinInteractor() *usecase.MockPenguinInteractor {
	return new(usecase.MockPenguinInteractor)
}

func TestPenguinCuiController_Quit(t *testing.T) {
	m := newMockPenguinInteractor()
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestPenguinCuiController_Reset(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("Reset").Return("reset_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestPenguinCuiController_GiveUp(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("GiveUp").Return("giveup_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestPenguinCuiController_Hint(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("Hint").Return("hint_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestPenguinCuiController_AutoComplete(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("AutoComplete").Return("ac_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestPenguinCuiController_ActionLog(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("ActionLog").Return("log_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestPenguinCuiController_Undo(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("Undo").Return("undo_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestPenguinCuiController_MoveTableauToFoundation(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("MoveTableauToFoundation", 2).Return("t2f_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "t2f_output", c.Exec("m t 2 f"))
}

func TestPenguinCuiController_MoveTableauToTableauWithCardIndex(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("MoveTableauToTableau", 0, 3, 4).Return("t2t_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "t2t_output", c.Exec("m t 0 3 t 4"))
}

func TestPenguinCuiController_MoveTableauToTableauTopCard(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("MoveTableauToTableau", 0, -1, 4).Return("t2t_top_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "t2t_top_output", c.Exec("m t 0 t 4"))
}

func TestPenguinCuiController_MoveTableauToFreeCell(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("MoveTableauToFreeCell", 0, 5).Return("t2c_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "t2c_output", c.Exec("m t 0 c 5"))
}

func TestPenguinCuiController_MoveFreeCellToTableau(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("MoveFreeCellToTableau", 1, 3).Return("c2t_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "c2t_output", c.Exec("m c 1 t 3"))
}

func TestPenguinCuiController_MoveFreeCellToFoundation(t *testing.T) {
	m := newMockPenguinInteractor()
	m.On("MoveFreeCellToFoundation", 0).Return("c2f_output")
	c := controller.NewPenguinCuiController(m)

	assert.Equal(t, "c2f_output", c.Exec("m c 0 f"))
}

func TestPenguinCuiController_MoveErrors(t *testing.T) {
	promptTests := []struct {
		name  string
		input string
	}{
		{"move no args", "m"},
		{"move tableau no col args", "m t"},
		{"move tableau to tableau no toCol", "m t 0 t"},
		{"move tableau to freecell no cell", "m t 0 c"},
		{"move freecell no args", "m c"},
		{"move freecell to tableau no col", "m c 0 t"},
		{"move tableau cardIdx t chained wizard toCol", "m t 0 3 t"},
	}

	for _, tt := range promptTests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockPenguinInteractor()
			c := controller.NewPenguinCuiController(m)
			result := c.Exec(tt.input)
			assert.True(t, cuiutil.IsPromptRequest(result))
		})
	}

	errorTests := []struct {
		name     string
		input    string
		contains string
	}{
		{"move invalid from zone", "m x t 3", "x"},
		{"move tableau invalid col", "m t abc f", "abc"},
		{"move tableau to tableau invalid toCol", "m t 0 t abc", "abc"},
		{"move tableau to invalid zone", "m t 0 x", "m t"},
		{"move tableau cardIndex wrong zone", "m t 0 3 f 4", "m t"},
		{"move tableau invalid cardIndex", "m t 0 abc t 4", "m t"},
		{"move tableau cardIndex invalid toCol", "m t 0 3 t abc", "abc"},
		{"move tableau to freecell invalid cell", "m t 0 c abc", "abc"},
		{"move freecell invalid cell", "m c abc t 3", "abc"},
		{"move freecell to tableau invalid col", "m c 0 t abc", "abc"},
		{"move freecell to invalid zone", "m c 0 x", "x"},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockPenguinInteractor()
			c := controller.NewPenguinCuiController(m)
			result := c.Exec(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestPenguinCuiController_MoveShorthand(t *testing.T) {
	t.Run("m <from> <to> moves top card", func(t *testing.T) {
		m := newMockPenguinInteractor()
		c := controller.NewPenguinCuiController(m)
		m.On("MoveTableauToTableau", 0, -1, 1).Return("move_output")
		assert.Equal(t, "move_output", c.Exec("m 0 1"))
	})
	t.Run("m <from> prompts for destination", func(t *testing.T) {
		m := newMockPenguinInteractor()
		c := controller.NewPenguinCuiController(m)
		result := c.Exec("m 0")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "m 0 {0}", tmpl)
	})
	t.Run("m <from> <invalid> returns error", func(t *testing.T) {
		m := newMockPenguinInteractor()
		c := controller.NewPenguinCuiController(m)
		result := c.Exec("m 0 abc")
		assert.Contains(t, result, "abc")
	})
}

func TestPenguinCuiController_FoundationShorthand(t *testing.T) {
	t.Run("f <col> moves tableau to foundation", func(t *testing.T) {
		m := newMockPenguinInteractor()
		c := controller.NewPenguinCuiController(m)
		m.On("MoveTableauToFoundation", 2).Return("tf_output")
		assert.Equal(t, "tf_output", c.Exec("f 2"))
	})
	t.Run("f with no args prompts for column", func(t *testing.T) {
		m := newMockPenguinInteractor()
		c := controller.NewPenguinCuiController(m)
		result := c.Exec("f")
		assert.True(t, cuiutil.IsPromptRequest(result))
		_, tmpl := cuiutil.ParsePromptRequest(result)
		assert.Equal(t, "f {0}", tmpl)
	})
	t.Run("f <invalid> returns error", func(t *testing.T) {
		m := newMockPenguinInteractor()
		c := controller.NewPenguinCuiController(m)
		result := c.Exec("f abc")
		assert.Contains(t, result, "abc")
	})
}

func TestPenguinCuiController_UnknownCommand(t *testing.T) {
	m := newMockPenguinInteractor()
	c := controller.NewPenguinCuiController(m)

	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

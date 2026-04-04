package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockFreeCellInteractor() *mockusecase.MockFreeCellInteractor {
	return new(mockusecase.MockFreeCellInteractor)
}

func TestFreeCellCuiController_Quit(t *testing.T) {
	m := newMockFreeCellInteractor()
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestFreeCellCuiController_Reset(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("Reset").Return("reset_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestFreeCellCuiController_GiveUp(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("GiveUp").Return("giveup_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestFreeCellCuiController_Hint(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("Hint").Return("hint_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestFreeCellCuiController_AutoComplete(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("AutoComplete").Return("ac_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestFreeCellCuiController_ActionLog(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("ActionLog").Return("log_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestFreeCellCuiController_Undo(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("Undo").Return("undo_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestFreeCellCuiController_MoveTableauToFoundation(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("MoveTableauToFoundation", 2).Return("t2f_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "t2f_output", c.Exec("m t 2 f"))
}

func TestFreeCellCuiController_MoveTableauToTableauWithCardIndex(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("MoveTableauToTableau", 0, 3, 4).Return("t2t_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "t2t_output", c.Exec("m t 0 3 t 4"))
}

func TestFreeCellCuiController_MoveTableauToTableauTopCard(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("MoveTableauToTableau", 0, -1, 4).Return("t2t_top_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "t2t_top_output", c.Exec("m t 0 t 4"))
}

func TestFreeCellCuiController_MoveTableauToFreeCell(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("MoveTableauToFreeCell", 0, 1).Return("t2c_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "t2c_output", c.Exec("m t 0 c 1"))
}

func TestFreeCellCuiController_MoveFreeCellToTableau(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("MoveFreeCellToTableau", 1, 3).Return("c2t_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "c2t_output", c.Exec("m c 1 t 3"))
}

func TestFreeCellCuiController_MoveFreeCellToFoundation(t *testing.T) {
	m := newMockFreeCellInteractor()
	m.On("MoveFreeCellToFoundation", 0).Return("c2f_output")
	c := NewFreeCellCuiController(m)

	assert.Equal(t, "c2f_output", c.Exec("m c 0 f"))
}

func TestFreeCellCuiController_MoveErrors(t *testing.T) {
	// Cases where args are completely missing now return prompt requests
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
	}

	for _, tt := range promptTests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockFreeCellInteractor()
			c := NewFreeCellCuiController(m)
			result := c.Exec(tt.input)
			assert.True(t, cuiutil.IsPromptRequest(result))
		})
	}

	// Cases where args are present but invalid still return error strings
	errorTests := []struct {
		name     string
		input    string
		contains string
	}{
		{"move invalid from zone", "m x t 3", "Invalid from zone"},
		{"move tableau invalid col", "m t abc f", "Invalid from column"},
		{"move tableau to tableau invalid toCol", "m t 0 t abc", "Invalid to column"},
		{"move tableau to invalid zone", "m t 0 x", "Invalid move command"},
		{"move tableau cardIndex wrong zone", "m t 0 3 f 4", "Invalid move command"},
		{"move tableau cardIndex too few args", "m t 0 3 t", "Invalid move command"},
		{"move tableau invalid cardIndex", "m t 0 abc t 4", "Invalid move command"},
		{"move tableau cardIndex invalid toCol", "m t 0 3 t abc", "Invalid to column"},
		{"move tableau to freecell invalid cell", "m t 0 c abc", "Invalid cell"},
		{"move freecell invalid cell", "m c abc t 3", "Invalid cell"},
		{"move freecell to tableau invalid col", "m c 0 t abc", "Invalid column"},
		{"move freecell to invalid zone", "m c 0 x", "Invalid to zone"},
		{"move invalid from zone w", "m w t 3", "Invalid from zone"},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMockFreeCellInteractor()
			c := NewFreeCellCuiController(m)
			result := c.Exec(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestFreeCellCuiController_UnknownCommand(t *testing.T) {
	m := newMockFreeCellInteractor()
	c := NewFreeCellCuiController(m)

	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
	assert.Contains(t, c.Exec(""), "'help' でコマンド一覧を表示します。")
}

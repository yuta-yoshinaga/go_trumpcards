package controller

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockScorpionInteractor() *mockusecase.MockScorpionInteractor {
	return new(mockusecase.MockScorpionInteractor)
}

func TestScorpionCuiControllerQuit(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestScorpionCuiControllerReset(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestScorpionCuiControllerDeal(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("Deal").Return("deal_output")
	assert.Equal(t, "deal_output", c.Exec("d"))
	assert.Equal(t, "deal_output", c.Exec("deal"))
}

func TestScorpionCuiControllerLegal(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("LegalMoves", 2).Return("legal_output")
	assert.Equal(t, "legal_output", c.Exec("legal 2"))
}

func TestScorpionCuiControllerLegalPrompt(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	result := c.Exec("legal")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestScorpionCuiControllerLegalInvalidCol(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	assert.NotEmpty(t, c.Exec("legal abc"))
}

func TestScorpionCuiControllerGiveUp(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestScorpionCuiControllerHint(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestScorpionCuiControllerAutoComplete(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("AutoComplete").Return("ac_output")
	assert.Equal(t, "ac_output", c.Exec("ac"))
	assert.Equal(t, "ac_output", c.Exec("autocomplete"))
}

func TestScorpionCuiControllerActionLog(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

// #6339: u / undo / un / undo_n に数値を渡すと UndoN(n) が呼ばれる。
func TestScorpionCuiControllerUndoTakesACount(t *testing.T) {
	for _, alias := range []string{"u", "undo", "un", "undo_n"} {
		t.Run(alias, func(t *testing.T) {
			si := newMockScorpionInteractor()
			c := NewScorpionCuiController(si)
			si.On("UndoN", 3).Return("undone 3")
			assert.Equal(t, "undone 3", c.Exec(alias+" 3"))
			si.AssertCalled(t, "UndoN", 3)
			si.AssertNotCalled(t, "Undo")
		})
	}
}

// 引数なしの u / undo は従来どおり Undo() を呼び、UndoN は呼ばない。
func TestScorpionCuiControllerBareUndoStillUndoesOnce(t *testing.T) {
	for _, alias := range []string{"u", "undo"} {
		t.Run(alias, func(t *testing.T) {
			si := newMockScorpionInteractor()
			c := NewScorpionCuiController(si)
			si.On("Undo").Return("undone")
			assert.Equal(t, "undone", c.Exec(alias))
			si.AssertCalled(t, "Undo")
			si.AssertNotCalled(t, "UndoN", mock.Anything)
			si.AssertNotCalled(t, "UndoToEscape")
		})
	}
}

// 引数なしの un / undo_n は UndoToEscape() が返した件数（0でも1でもない値）を UndoN に渡す。
func TestScorpionCuiControllerBareUnUsesUndoToEscape(t *testing.T) {
	for _, alias := range []string{"un", "undo_n"} {
		t.Run(alias, func(t *testing.T) {
			si := newMockScorpionInteractor()
			c := NewScorpionCuiController(si)
			si.On("UndoToEscape").Return(4)
			si.On("UndoN", 4).Return("undone 4")
			assert.Equal(t, "undone 4", c.Exec(alias))
			si.AssertCalled(t, "UndoToEscape")
			si.AssertCalled(t, "UndoN", 4)
			si.AssertNotCalled(t, "Undo")
		})
	}
}

// UndoToEscape() が 0 以下のとき、UndoN は呼ばれず、メッセージが返る。
func TestScorpionCuiControllerBareUnWhenNoEscape(t *testing.T) {
	for _, val := range []int{0, -1} {
		t.Run(strconv.Itoa(val), func(t *testing.T) {
			si := newMockScorpionInteractor()
			c := NewScorpionCuiController(si)
			si.On("UndoToEscape").Return(val)
			out := c.Exec("un")
			assert.Contains(t, out, i18n.T("scorpion.noUndoToEscape"))
			si.AssertCalled(t, "UndoToEscape")
			si.AssertNotCalled(t, "UndoN", mock.Anything)
			si.AssertNotCalled(t, "Undo")
		})
	}
}

// 不正な引数（0, -1, 数値以外）はエラーメッセージを返し、Undo / UndoN は呼ばない。
func TestScorpionCuiControllerUndoRejectsABadCount(t *testing.T) {
	for _, arg := range []string{"0", "-1", "zz", "1.5"} {
		t.Run(arg, func(t *testing.T) {
			si := newMockScorpionInteractor()
			c := NewScorpionCuiController(si)
			out := c.Exec("un " + arg)
			assert.Contains(t, out, i18n.Tf("scorpion.invalidUndoCount", "val", arg))
			si.AssertNotCalled(t, "UndoN", mock.Anything)
			si.AssertNotCalled(t, "Undo")
			si.AssertNotCalled(t, "UndoToEscape")
		})
	}
}

// 回数の上限は決めずに素通しする。
func TestScorpionCuiControllerUndoPassesLargeCountsThrough(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("UndoN", 9999).Return("not enough history")
	assert.Equal(t, "not enough history", c.Exec("un 9999"))
	si.AssertCalled(t, "UndoN", 9999)
}

func TestScorpionCuiControllerMoveShorthandTopCard(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("MoveTableauToTableau", 0, -1, 3).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 3"))
}

func TestScorpionCuiControllerMoveShorthandWithIdx(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("MoveTableauToTableau", 0, 2, 4).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 0 2 4"))
}

func TestScorpionCuiControllerMoveShorthandPrompt(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	result := c.Exec("m 0")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestScorpionCuiControllerMovePrompt(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	result := c.Exec("m")
	assert.Contains(t, result, cuiutil.PromptPrefix)
}

func TestScorpionCuiControllerMoveTableauToTableau(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)
	si.On("MoveTableauToTableau", 0, 2, 4).Return("move_tt_output")
	assert.Equal(t, "move_tt_output", c.Exec("m t 0 2 t 4"))
}

func TestScorpionCuiControllerMoveTableauPrompts(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)

	assert.Contains(t, c.Exec("m t"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m t 0"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m t 0 2 t"), cuiutil.PromptPrefix)
}

func TestScorpionCuiControllerMoveInvalidFrom(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)

	result := c.Exec("m w")
	assert.NotEmpty(t, result)
}

func TestScorpionCuiControllerMoveUsage(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)

	result := c.Exec("m t 0 2 x")
	assert.NotEmpty(t, result)
}

func TestScorpionCuiControllerMoveInvalidCol(t *testing.T) {
	si := newMockScorpionInteractor()
	c := NewScorpionCuiController(si)

	assert.NotEmpty(t, c.Exec("m t abc"))
	assert.NotEmpty(t, c.Exec("m 0 abc"))
	assert.NotEmpty(t, c.Exec("m 0 1 abc"))
	assert.NotEmpty(t, c.Exec("m t 0 abc t 4"))
	assert.NotEmpty(t, c.Exec("m t 0 1 t abc"))
}

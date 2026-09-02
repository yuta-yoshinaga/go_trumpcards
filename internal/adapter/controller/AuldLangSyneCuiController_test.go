package controller

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockAuldLangSyneInteractor() *mockusecase.MockAuldLangSyneInteractor {
	return new(mockusecase.MockAuldLangSyneInteractor)
}

func TestAuldLangSyneCuiController_Quit(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestAuldLangSyneCuiController_Reset(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("Reset").Return("reset_out")
	assert.Equal(t, "reset_out", c.Exec("r"))
	assert.Equal(t, "reset_out", c.Exec("reset"))
}

// Deal is this game's own command, in the slot where Sir Tommy has `s w <idx>`.
func TestAuldLangSyneCuiController_Deal(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("Deal").Return("deal_out")
	assert.Equal(t, "deal_out", c.Exec("d"))
	assert.Equal(t, "deal_out", c.Exec("deal"))
}

func TestAuldLangSyneCuiController_GiveUp(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("GiveUp").Return("giveup_out")
	assert.Equal(t, "giveup_out", c.Exec("g"))
	assert.Equal(t, "giveup_out", c.Exec("giveup"))
}

func TestAuldLangSyneCuiController_AutoComplete(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("AutoComplete").Return("auto_out")
	assert.Equal(t, "auto_out", c.Exec("ac"))
	assert.Equal(t, "auto_out", c.Exec("autocomplete"))
}

func TestAuldLangSyneCuiController_Undo(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("Undo").Return("undo_out")
	assert.Equal(t, "undo_out", c.Exec("u"))
	assert.Equal(t, "undo_out", c.Exec("undo"))
}

// u / undo / un / undo_n に数値を渡すと UndoN(n) が呼ばれる。
func TestAuldLangSyneCuiController_UndoTakesACount(t *testing.T) {
	for _, alias := range []string{"u", "undo", "un", "undo_n"} {
		t.Run(alias, func(t *testing.T) {
			ci := newMockAuldLangSyneInteractor()
			c := NewAuldLangSyneCuiController(ci)
			ci.On("UndoN", 3).Return("undone 3")
			assert.Equal(t, "undone 3", c.Exec(alias+" 3"))
			ci.AssertCalled(t, "UndoN", 3)
			ci.AssertNotCalled(t, "Undo")
		})
	}
}

// 引数なしの u / undo は従来どおり Undo() を呼び、UndoN は呼ばない。
func TestAuldLangSyneCuiController_BareUndoStillUndoesOnce(t *testing.T) {
	for _, alias := range []string{"u", "undo"} {
		t.Run(alias, func(t *testing.T) {
			ci := newMockAuldLangSyneInteractor()
			c := NewAuldLangSyneCuiController(ci)
			ci.On("Undo").Return("undone")
			assert.Equal(t, "undone", c.Exec(alias))
			ci.AssertCalled(t, "Undo")
			ci.AssertNotCalled(t, "UndoN", mock.Anything)
			ci.AssertNotCalled(t, "UndoToEscape")
		})
	}
}

// 引数なしの un / undo_n は UndoToEscape() が返した件数を UndoN に渡す。
func TestAuldLangSyneCuiController_BareUnUsesUndoToEscape(t *testing.T) {
	for _, alias := range []string{"un", "undo_n"} {
		t.Run(alias, func(t *testing.T) {
			ci := newMockAuldLangSyneInteractor()
			c := NewAuldLangSyneCuiController(ci)
			ci.On("UndoToEscape").Return(4)
			ci.On("UndoN", 4).Return("undone 4")
			assert.Equal(t, "undone 4", c.Exec(alias))
			ci.AssertCalled(t, "UndoToEscape")
			ci.AssertCalled(t, "UndoN", 4)
			ci.AssertNotCalled(t, "Undo")
		})
	}
}

// UndoToEscape() が 0 以下のとき、UndoN は呼ばれず、メッセージが返る。
func TestAuldLangSyneCuiController_BareUnWhenNoEscape(t *testing.T) {
	for _, val := range []int{0, -1} {
		t.Run(strconv.Itoa(val), func(t *testing.T) {
			ci := newMockAuldLangSyneInteractor()
			c := NewAuldLangSyneCuiController(ci)
			ci.On("UndoToEscape").Return(val)
			out := c.Exec("un")
			assert.Contains(t, out, "手詰まり状態ではないか、脱出可能なアンドゥがありません")
			ci.AssertCalled(t, "UndoToEscape")
			ci.AssertNotCalled(t, "UndoN", mock.Anything)
			ci.AssertNotCalled(t, "Undo")
		})
	}
}

// 不正な引数（0, -1, 数値以外）はエラーメッセージを返し、Undo / UndoN は呼ばない。
func TestAuldLangSyneCuiController_UndoRejectsABadCount(t *testing.T) {
	for _, arg := range []string{"0", "-1", "zz", "1.5"} {
		t.Run(arg, func(t *testing.T) {
			ci := newMockAuldLangSyneInteractor()
			c := NewAuldLangSyneCuiController(ci)
			out := c.Exec("un " + arg)
			assert.Contains(t, out, "無効なアンドゥ回数です: "+arg+"（1以上の整数を指定してください）")
			ci.AssertNotCalled(t, "UndoN", mock.Anything)
			ci.AssertNotCalled(t, "Undo")
			ci.AssertNotCalled(t, "UndoToEscape")
		})
	}
}

// 回数の上限は決めずに素通しする。
func TestAuldLangSyneCuiController_UndoPassesLargeCountsThrough(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("UndoN", 9999).Return("not enough history")
	assert.Equal(t, "not enough history", c.Exec("un 9999"))
	ci.AssertCalled(t, "UndoN", 9999)
}

func TestAuldLangSyneCuiController_Hint(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("Hint").Return("hint_out")
	assert.Equal(t, "hint_out", c.Exec("h"))
	assert.Equal(t, "hint_out", c.Exec("hint"))
}

func TestAuldLangSyneCuiController_ActionLog(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("ActionLog").Return("log_out")
	assert.Equal(t, "log_out", c.Exec("l"))
	assert.Equal(t, "log_out", c.Exec("log"))
}

func TestAuldLangSyneCuiController_WasteToFoundation(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	ci.On("PlayWasteToFoundation", 3, 0).Return("ok")
	assert.Equal(t, "ok", c.Exec("w 3 f 0"))
}

func TestAuldLangSyneCuiController_WasteMove_Prompts(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	assert.Contains(t, c.Exec("w"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("w 2"), cuiutil.PromptPrefix)
	// `f` missing from the third slot is the same "incomplete, ask for the rest"
	// case rather than an error.
	assert.Contains(t, c.Exec("w 2 x 1"), cuiutil.PromptPrefix)
}

func TestAuldLangSyneCuiController_WasteMove_Invalid(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	assert.NotEmpty(t, c.Exec("w abc f 1"))
	assert.NotEmpty(t, c.Exec("w 1 f abc"))
	ci.AssertNotCalled(t, "PlayWasteToFoundation")
}

func TestAuldLangSyneCuiController_Unknown(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	assert.NotEmpty(t, c.Exec("unknowncmd"))
}

func TestAuldLangSyneCuiController_Empty(t *testing.T) {
	ci := newMockAuldLangSyneInteractor()
	c := NewAuldLangSyneCuiController(ci)
	assert.NotEmpty(t, c.Exec(""))
}

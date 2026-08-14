package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockSirTommyInteractor() *mockusecase.MockSirTommyInteractor {
	return new(mockusecase.MockSirTommyInteractor)
}

func TestSirTommyCuiController_Quit(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestSirTommyCuiController_Reset(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("Reset").Return("reset_out")
	assert.Equal(t, "reset_out", c.Exec("r"))
	assert.Equal(t, "reset_out", c.Exec("reset"))
}

func TestSirTommyCuiController_GiveUp(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("GiveUp").Return("giveup_out")
	assert.Equal(t, "giveup_out", c.Exec("g"))
	assert.Equal(t, "giveup_out", c.Exec("giveup"))
}

func TestSirTommyCuiController_AutoComplete(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("AutoComplete").Return("auto_out")
	assert.Equal(t, "auto_out", c.Exec("ac"))
	assert.Equal(t, "auto_out", c.Exec("autocomplete"))
}

func TestSirTommyCuiController_Undo(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("Undo").Return("undo_out")
	assert.Equal(t, "undo_out", c.Exec("u"))
}

func TestSirTommyCuiController_Hint(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("Hint").Return("hint_out")
	assert.Equal(t, "hint_out", c.Exec("h"))
	assert.Equal(t, "hint_out", c.Exec("hint"))
}

func TestSirTommyCuiController_ActionLog(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("ActionLog").Return("log_out")
	assert.Equal(t, "log_out", c.Exec("l"))
	assert.Equal(t, "log_out", c.Exec("log"))
}

func TestSirTommyCuiController_StockToFoundation(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("PlayStockToFoundation", 1).Return("ok")
	assert.Equal(t, "ok", c.Exec("s f 1"))
}

func TestSirTommyCuiController_StockToWaste(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("PlayStockToWaste", 2).Return("ok")
	assert.Equal(t, "ok", c.Exec("s w 2"))
}

func TestSirTommyCuiController_WasteToFoundation(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	ci.On("PlayWasteToFoundation", 3, 0).Return("ok")
	assert.Equal(t, "ok", c.Exec("w 3 f 0"))
}

func TestSirTommyCuiController_StockMove_Prompts(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	assert.Contains(t, c.Exec("s"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("s f"), cuiutil.PromptPrefix)
}

func TestSirTommyCuiController_StockMove_InvalidDest(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	out := c.Exec("s x 1")
	assert.NotEmpty(t, out)
}

func TestSirTommyCuiController_StockMove_InvalidIdx(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	assert.NotEmpty(t, c.Exec("s f abc"))
}

func TestSirTommyCuiController_WasteMove_Prompts(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	assert.Contains(t, c.Exec("w"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("w 2"), cuiutil.PromptPrefix)
}

func TestSirTommyCuiController_WasteMove_Invalid(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	assert.NotEmpty(t, c.Exec("w abc f 1"))
	assert.NotEmpty(t, c.Exec("w 1 f abc"))
}

func TestSirTommyCuiController_Unknown(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	assert.NotEmpty(t, c.Exec("unknowncmd"))
}

func TestSirTommyCuiController_Empty(t *testing.T) {
	ci := newMockSirTommyInteractor()
	c := NewSirTommyCuiController(ci)
	assert.NotEmpty(t, c.Exec(""))
}

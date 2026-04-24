package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockCalculationInteractor() *mockusecase.MockCalculationInteractor {
	return new(mockusecase.MockCalculationInteractor)
}

func TestCalculationCuiController_Quit(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestCalculationCuiController_Reset(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("Reset").Return("reset_out")
	assert.Equal(t, "reset_out", c.Exec("r"))
	assert.Equal(t, "reset_out", c.Exec("reset"))
}

func TestCalculationCuiController_GiveUp(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("GiveUp").Return("giveup_out")
	assert.Equal(t, "giveup_out", c.Exec("g"))
	assert.Equal(t, "giveup_out", c.Exec("giveup"))
}

func TestCalculationCuiController_AutoComplete(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("AutoComplete").Return("auto_out")
	assert.Equal(t, "auto_out", c.Exec("ac"))
	assert.Equal(t, "auto_out", c.Exec("autocomplete"))
}

func TestCalculationCuiController_Undo(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("Undo").Return("undo_out")
	assert.Equal(t, "undo_out", c.Exec("u"))
}

func TestCalculationCuiController_Hint(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("Hint").Return("hint_out")
	assert.Equal(t, "hint_out", c.Exec("h"))
	assert.Equal(t, "hint_out", c.Exec("hint"))
}

func TestCalculationCuiController_ActionLog(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("ActionLog").Return("log_out")
	assert.Equal(t, "log_out", c.Exec("l"))
	assert.Equal(t, "log_out", c.Exec("log"))
}

func TestCalculationCuiController_StockToFoundation(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("PlayStockToFoundation", 1).Return("ok")
	assert.Equal(t, "ok", c.Exec("s f 1"))
}

func TestCalculationCuiController_StockToWaste(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("PlayStockToWaste", 2).Return("ok")
	assert.Equal(t, "ok", c.Exec("s w 2"))
}

func TestCalculationCuiController_WasteToFoundation(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	ci.On("PlayWasteToFoundation", 3, 0).Return("ok")
	assert.Equal(t, "ok", c.Exec("w 3 f 0"))
}

func TestCalculationCuiController_StockMove_Prompts(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	assert.Contains(t, c.Exec("s"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("s f"), cuiutil.PromptPrefix)
}

func TestCalculationCuiController_StockMove_InvalidDest(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	out := c.Exec("s x 1")
	assert.NotEmpty(t, out)
}

func TestCalculationCuiController_StockMove_InvalidIdx(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	assert.NotEmpty(t, c.Exec("s f abc"))
}

func TestCalculationCuiController_WasteMove_Prompts(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	assert.Contains(t, c.Exec("w"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("w 2"), cuiutil.PromptPrefix)
}

func TestCalculationCuiController_WasteMove_Invalid(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	assert.NotEmpty(t, c.Exec("w abc f 1"))
	assert.NotEmpty(t, c.Exec("w 1 f abc"))
}

func TestCalculationCuiController_Unknown(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	assert.NotEmpty(t, c.Exec("unknowncmd"))
}

func TestCalculationCuiController_Empty(t *testing.T) {
	ci := newMockCalculationInteractor()
	c := NewCalculationCuiController(ci)
	assert.NotEmpty(t, c.Exec(""))
}

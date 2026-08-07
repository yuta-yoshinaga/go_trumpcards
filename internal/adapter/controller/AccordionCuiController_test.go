package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
)

func newMockAccordionInteractor() *mockusecase.MockAccordionInteractor {
	return new(mockusecase.MockAccordionInteractor)
}

func TestAccordionCuiControllerQuit(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestAccordionCuiControllerReset(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	ai.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestAccordionCuiControllerGiveUp(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	ai.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestAccordionCuiControllerHint(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	ai.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestAccordionCuiControllerActionLog(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	ai.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestAccordionCuiControllerUndo(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	ai.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestAccordionCuiControllerMove(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	ai.On("Move", 3, 0).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("m 3 0"))
}

func TestAccordionCuiControllerMovePrompt(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	assert.Contains(t, c.Exec("m"), cuiutil.PromptPrefix)
	assert.Contains(t, c.Exec("m 3"), cuiutil.PromptPrefix)
}

func TestAccordionCuiControllerMoveInvalid(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	assert.NotEmpty(t, c.Exec("m abc"))
	assert.NotEmpty(t, c.Exec("m 3 xyz"))
}

func TestAccordionCuiControllerUnknown(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	result := c.Exec("unknowncmd")
	assert.NotEmpty(t, result)
}

func TestAccordionCuiControllerEmpty(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	result := c.Exec("")
	assert.NotEmpty(t, result)
}

// **Web は1クリックで最後まで自動化できるのに、ネイティブ CUI には
// オートコンプリートが無かった (#4793)。**姉妹の Wasp は ac/autocomplete を
// 公開している。
func TestAccordionCuiControllerAutoComplete(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	ai.On("AutoComplete").Return("auto_output")

	assert.Equal(t, "auto_output", c.Exec("ac"))
	assert.Equal(t, "auto_output", c.Exec("autocomplete"))
}

// **既存コマンドは何も変わらない。**
func TestAccordionCuiControllerAutoCompleteDoesNotShadowOthers(t *testing.T) {
	ai := newMockAccordionInteractor()
	c := NewAccordionCuiController(ai)
	ai.On("GiveUp").Return("giveup_output")

	assert.Equal(t, "giveup_output", c.Exec("g"))
	ai.AssertNotCalled(t, "AutoComplete")
}

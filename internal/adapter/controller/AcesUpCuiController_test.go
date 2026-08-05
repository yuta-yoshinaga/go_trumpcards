package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockAcesUpInteractor() *mockusecase.MockAcesUpInteractor {
	return new(mockusecase.MockAcesUpInteractor)
}

func TestAcesUpCuiControllerQuit(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestAcesUpCuiControllerReset(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestAcesUpCuiControllerDraw(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestAcesUpCuiControllerGiveUp(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestAcesUpCuiControllerHint(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestAcesUpCuiControllerUndo(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestAcesUpCuiControllerActionLog(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestAcesUpCuiControllerRemove(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Remove", 2).Return("remove_output")
	assert.Equal(t, "remove_output", c.Exec("rm 2"))
	assert.Equal(t, "remove_output", c.Exec("remove 2"))
}

func TestAcesUpCuiControllerMove(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	ai.On("Move", 1).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("mv 1"))
	assert.Equal(t, "move_output", c.Exec("move 1"))
}

func TestAcesUpCuiControllerColCommand_InvalidArgs(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	// **エラーも日本語ロケールでは日本語。**ここだけ英語リテラルを返していて、
	// このテストがその挙動を固定していた (#4803)。
	assert.Contains(t, c.Exec("rm"), "使い方: rm")
	assert.Contains(t, c.Exec("rm 3 0"), "使い方: rm")
	assert.Contains(t, c.Exec("rm a"), "無効な列番号です: a")
	assert.Contains(t, c.Exec("mv"), "使い方: mv")
	assert.Contains(t, c.Exec("mv x"), "無効な列番号です: x")
	assert.NotContains(t, c.Exec("rm a"), "Invalid col")

	// 英語ロケールでは従来と同じ文面 (受け入れ条件2)。
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	assert.Contains(t, c.Exec("rm"), "Usage: rm <col>")
	assert.Contains(t, c.Exec("rm a"), "Invalid column: a")
}

func TestAcesUpCuiControllerUnknown(t *testing.T) {
	ai := newMockAcesUpInteractor()
	c := NewAcesUpCuiController(ai)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

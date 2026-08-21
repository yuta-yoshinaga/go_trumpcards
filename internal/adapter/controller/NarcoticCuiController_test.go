package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func newMockNarcoticInteractor() *mockusecase.MockNarcoticInteractor {
	return new(mockusecase.MockNarcoticInteractor)
}

func TestNarcoticCuiControllerQuit(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	assert.Equal(t, "bye.", c.Exec("q"))
	assert.Equal(t, "bye.", c.Exec("quit"))
}

func TestNarcoticCuiControllerReset(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	ai.On("Reset").Return("reset_output")
	assert.Equal(t, "reset_output", c.Exec("r"))
	assert.Equal(t, "reset_output", c.Exec("reset"))
}

func TestNarcoticCuiControllerDraw(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	ai.On("Draw").Return("draw_output")
	assert.Equal(t, "draw_output", c.Exec("d"))
	assert.Equal(t, "draw_output", c.Exec("draw"))
}

func TestNarcoticCuiControllerGiveUp(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	ai.On("GiveUp").Return("giveup_output")
	assert.Equal(t, "giveup_output", c.Exec("g"))
	assert.Equal(t, "giveup_output", c.Exec("giveup"))
}

func TestNarcoticCuiControllerHint(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	ai.On("Hint").Return("hint_output")
	assert.Equal(t, "hint_output", c.Exec("h"))
	assert.Equal(t, "hint_output", c.Exec("hint"))
}

func TestNarcoticCuiControllerUndo(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	ai.On("Undo").Return("undo_output")
	assert.Equal(t, "undo_output", c.Exec("u"))
	assert.Equal(t, "undo_output", c.Exec("undo"))
}

func TestNarcoticCuiControllerActionLog(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	ai.On("ActionLog").Return("log_output")
	assert.Equal(t, "log_output", c.Exec("log"))
	assert.Equal(t, "log_output", c.Exec("l"))
}

func TestNarcoticCuiControllerRemove(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	// **列を取らない。**揃った4枚をまとめて捨てるので、選ぶ余地が無い。
	// クローン元 (Aces Up) は列ごとに捨てるので `rm <col>` だった。
	ai.On("Remove").Return("remove_output")
	assert.Equal(t, "remove_output", c.Exec("rm"))
	assert.Equal(t, "remove_output", c.Exec("remove"))
	ai.AssertCalled(t, "Remove")
}

func TestNarcoticCuiControllerRedeal(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	ai.On("Redeal").Return("redeal_output")
	assert.Equal(t, "redeal_output", c.Exec("rd"))
	assert.Equal(t, "redeal_output", c.Exec("redeal"))
}

func TestNarcoticCuiControllerMove(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	ai.On("Move", 1).Return("move_output")
	assert.Equal(t, "move_output", c.Exec("mv 1"))
	assert.Equal(t, "move_output", c.Exec("move 1"))
}

func TestNarcoticCuiControllerColCommand_InvalidArgs(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	// **エラーも日本語ロケールでは日本語。**ここだけ英語リテラルを返していて、
	// このテストがその挙動を固定していた (#4803)。
	assert.Contains(t, c.Exec("mv"), "使い方: mv")
	assert.Contains(t, c.Exec("mv x"), "無効な列番号です: x")
	assert.NotContains(t, c.Exec("mv x"), "Invalid col")

	// 英語ロケールでは従来と同じ文面 (受け入れ条件2)。
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	assert.Contains(t, c.Exec("mv"), "Usage: mv <col>")
	assert.Contains(t, c.Exec("mv x"), "Invalid column: x")
}

func TestNarcoticCuiControllerUnknown(t *testing.T) {
	ai := newMockNarcoticInteractor()
	c := NewNarcoticCuiController(ai)
	assert.Contains(t, c.Exec("xyz"), "コマンドが不明です")
}

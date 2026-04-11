package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockCanfieldGame() *interfaces.MockCanfieldGame {
	return new(interfaces.MockCanfieldGame)
}

func newMockCanfieldPresenter() *presenter.MockCanfieldPresenter {
	return new(presenter.MockCanfieldPresenter)
}

func TestNewCanfieldInteractor(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	assert.NotNil(t, ci)
}

func TestNewCanfieldInteractorPanicsOnNil(t *testing.T) {
	cp := newMockCanfieldPresenter()
	assert.Panics(t, func() { NewCanfieldInteractor(nil, cp) })
	cg := newMockCanfieldGame()
	assert.Panics(t, func() { NewCanfieldInteractor(cg, nil) })
}

func TestCanfieldInteractorReset(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("Reset").Return()
	cp.On("Output", cg, nil).Return("reset_output")
	assert.Equal(t, "reset_output", ci.Reset())
}

func TestCanfieldInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockCanfieldGame()
		cp := newMockCanfieldPresenter()
		ci := NewCanfieldInteractor(cg, cp)
		cg.On("Draw").Return(nil)
		cp.On("Output", cg, nil).Return("draw_output")
		assert.Equal(t, "draw_output", ci.Draw())
	})
	t.Run("error", func(t *testing.T) {
		cg := newMockCanfieldGame()
		cp := newMockCanfieldPresenter()
		ci := NewCanfieldInteractor(cg, cp)
		err := errors.New("no cards")
		cg.On("Draw").Return(err)
		cp.On("Output", cg, err).Return("error_output")
		assert.Equal(t, "error_output", ci.Draw())
	})
}

func TestCanfieldInteractorMoveWasteToTableau(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("MoveWasteToTableau", 2).Return(nil)
	cp.On("Output", cg, nil).Return("ok")
	assert.Equal(t, "ok", ci.MoveWasteToTableau(2))
}

func TestCanfieldInteractorMoveWasteToFoundation(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("MoveWasteToFoundation").Return(nil)
	cp.On("Output", cg, nil).Return("ok")
	assert.Equal(t, "ok", ci.MoveWasteToFoundation())
}

func TestCanfieldInteractorMoveTableauToTableau(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("MoveTableauToTableau", 0, 1, 2).Return(nil)
	cp.On("Output", cg, nil).Return("ok")
	assert.Equal(t, "ok", ci.MoveTableauToTableau(0, 1, 2))
}

func TestCanfieldInteractorMoveTableauToFoundation(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("MoveTableauToFoundation", 1).Return(nil)
	cp.On("Output", cg, nil).Return("ok")
	assert.Equal(t, "ok", ci.MoveTableauToFoundation(1))
}

func TestCanfieldInteractorMoveReserveToTableau(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("MoveReserveToTableau", 0).Return(nil)
	cp.On("Output", cg, nil).Return("ok")
	assert.Equal(t, "ok", ci.MoveReserveToTableau(0))
}

func TestCanfieldInteractorMoveReserveToFoundation(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("MoveReserveToFoundation").Return(nil)
	cp.On("Output", cg, nil).Return("ok")
	assert.Equal(t, "ok", ci.MoveReserveToFoundation())
}

func TestCanfieldInteractorGiveUp(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("GiveUp").Return()
	cp.On("Output", cg, nil).Return("giveup")
	assert.Equal(t, "giveup", ci.GiveUp())
}

func TestCanfieldInteractorHint(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cp.On("HintOutput", mock.Anything).Return("hint")
	assert.Equal(t, "hint", ci.Hint())
}

func TestCanfieldInteractorAutoComplete(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("AutoComplete").Return(nil)
	cp.On("Output", cg, nil).Return("ac")
	assert.Equal(t, "ac", ci.AutoComplete())
}

func TestCanfieldInteractorActionLog(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cp.On("ActionLogOutput", mock.Anything).Return("log")
	assert.Equal(t, "log", ci.ActionLog())
}

func TestCanfieldInteractorUndo(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("Undo").Return(nil)
	cp.On("Output", cg, nil).Return("undo")
	assert.Equal(t, "undo", ci.Undo())
}

func TestCanfieldInteractorUndoN(t *testing.T) {
	cg := newMockCanfieldGame()
	cp := newMockCanfieldPresenter()
	ci := NewCanfieldInteractor(cg, cp)
	cg.On("UndoN", 3).Return(nil)
	cp.On("Output", cg, nil).Return("undon")
	assert.Equal(t, "undon", ci.UndoN(3))
}

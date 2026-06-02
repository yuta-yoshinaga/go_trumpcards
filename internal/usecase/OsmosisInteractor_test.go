package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockOsmosisGame() *interfaces.MockOsmosisGame {
	return new(interfaces.MockOsmosisGame)
}

func newMockOsmosisPresenter() *presenter.MockOsmosisPresenter {
	return new(presenter.MockOsmosisPresenter)
}

func TestNewOsmosisInteractor(t *testing.T) {
	oi := NewOsmosisInteractor(newMockOsmosisGame(), newMockOsmosisPresenter())
	assert.NotNil(t, oi)
}

func TestNewOsmosisInteractorPanicsOnNil(t *testing.T) {
	op := newMockOsmosisPresenter()
	assert.Panics(t, func() { NewOsmosisInteractor(nil, op) })
	og := newMockOsmosisGame()
	assert.Panics(t, func() { NewOsmosisInteractor(og, nil) })
}

func TestOsmosisInteractorReset(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	og.On("Reset").Return()
	op.On("Output", og, nil).Return("reset_output")
	assert.Equal(t, "reset_output", oi.Reset())
}

func TestOsmosisInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		og := newMockOsmosisGame()
		op := newMockOsmosisPresenter()
		oi := NewOsmosisInteractor(og, op)
		og.On("Draw").Return(nil)
		op.On("Output", og, nil).Return("draw_output")
		assert.Equal(t, "draw_output", oi.Draw())
	})
	t.Run("error", func(t *testing.T) {
		og := newMockOsmosisGame()
		op := newMockOsmosisPresenter()
		oi := NewOsmosisInteractor(og, op)
		err := errors.New("no cards")
		og.On("Draw").Return(err)
		op.On("Output", og, err).Return("error_output")
		assert.Equal(t, "error_output", oi.Draw())
	})
}

func TestOsmosisInteractorMoveWasteToFoundation(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	og.On("MoveWasteToFoundation", 2).Return(nil)
	op.On("Output", og, nil).Return("ok")
	assert.Equal(t, "ok", oi.MoveWasteToFoundation(2))
}

func TestOsmosisInteractorMoveReserveToFoundation(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	og.On("MoveReserveToFoundation", 1, 3).Return(nil)
	op.On("Output", og, nil).Return("ok")
	assert.Equal(t, "ok", oi.MoveReserveToFoundation(1, 3))
}

func TestOsmosisInteractorGiveUp(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	og.On("GiveUp").Return()
	op.On("Output", og, nil).Return("giveup")
	assert.Equal(t, "giveup", oi.GiveUp())
}

func TestOsmosisInteractorHint(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	op.On("HintOutput", mock.Anything).Return("hint")
	assert.Equal(t, "hint", oi.Hint())
}

func TestOsmosisInteractorAutoComplete(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	og.On("AutoComplete").Return(nil)
	op.On("Output", og, nil).Return("ac")
	assert.Equal(t, "ac", oi.AutoComplete())
}

func TestOsmosisInteractorActionLog(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	op.On("ActionLogOutput", mock.Anything).Return("log")
	assert.Equal(t, "log", oi.ActionLog())
}

func TestOsmosisInteractorUndo(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	og.On("Undo").Return(nil)
	op.On("Output", og, nil).Return("undo")
	assert.Equal(t, "undo", oi.Undo())
}

func TestOsmosisInteractorUndoN(t *testing.T) {
	og := newMockOsmosisGame()
	op := newMockOsmosisPresenter()
	oi := NewOsmosisInteractor(og, op)
	og.On("UndoN", 3).Return(nil)
	op.On("Output", og, nil).Return("undon")
	assert.Equal(t, "undon", oi.UndoN(3))
}

func TestRestoreOsmosisInteractor(t *testing.T) {
	src := domain.NewDefaultOsmosis()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	oi, err := RestoreOsmosisInteractor(data, newMockOsmosisPresenter())
	assert.NoError(t, err)
	assert.NotNil(t, oi)

	_, err = RestoreOsmosisInteractor([]byte("not json"), newMockOsmosisPresenter())
	assert.Error(t, err)
}

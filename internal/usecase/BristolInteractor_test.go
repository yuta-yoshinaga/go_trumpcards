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

func newMockBristolGame() *interfaces.MockBristolGame {
	return new(interfaces.MockBristolGame)
}

func newMockBristolPresenter() *presenter.MockBristolPresenter {
	return new(presenter.MockBristolPresenter)
}

func TestNewBristolInteractor(t *testing.T) {
	bi := NewBristolInteractor(newMockBristolGame(), newMockBristolPresenter())
	assert.NotNil(t, bi)
}

func TestNewBristolInteractorPanicsOnNil(t *testing.T) {
	op := newMockBristolPresenter()
	assert.Panics(t, func() { NewBristolInteractor(nil, op) })
	bg := newMockBristolGame()
	assert.Panics(t, func() { NewBristolInteractor(bg, nil) })
}

func TestBristolInteractorReset(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("Reset").Return()
	op.On("Output", bg, nil).Return("reset_output")
	assert.Equal(t, "reset_output", bi.Reset())
}

func TestBristolInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBristolGame()
		op := newMockBristolPresenter()
		bi := NewBristolInteractor(bg, op)
		bg.On("Draw").Return(nil)
		op.On("Output", bg, nil).Return("draw_output")
		assert.Equal(t, "draw_output", bi.Draw())
	})
	t.Run("error", func(t *testing.T) {
		bg := newMockBristolGame()
		op := newMockBristolPresenter()
		bi := NewBristolInteractor(bg, op)
		err := errors.New("no cards")
		bg.On("Draw").Return(err)
		op.On("Output", bg, err).Return("error_output")
		assert.Equal(t, "error_output", bi.Draw())
	})
}

func TestBristolInteractorMoveTableauToTableau(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("MoveTableauToTableau", 0, 1).Return(nil)
	op.On("Output", bg, nil).Return("ok")
	assert.Equal(t, "ok", bi.MoveTableauToTableau(0, 1))
}

func TestBristolInteractorMoveTableauToFoundation(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("MoveTableauToFoundation", 2).Return(nil)
	op.On("Output", bg, nil).Return("ok")
	assert.Equal(t, "ok", bi.MoveTableauToFoundation(2))
}

func TestBristolInteractorMoveFanToTableau(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("MoveFanToTableau", 1, 3).Return(nil)
	op.On("Output", bg, nil).Return("ok")
	assert.Equal(t, "ok", bi.MoveFanToTableau(1, 3))
}

func TestBristolInteractorMoveFanToFoundation(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("MoveFanToFoundation", 0).Return(nil)
	op.On("Output", bg, nil).Return("ok")
	assert.Equal(t, "ok", bi.MoveFanToFoundation(0))
}

func TestBristolInteractorGiveUp(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("GiveUp").Return()
	op.On("Output", bg, nil).Return("giveup")
	assert.Equal(t, "giveup", bi.GiveUp())
}

func TestBristolInteractorHint(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	op.On("HintOutput", mock.Anything).Return("hint")
	assert.Equal(t, "hint", bi.Hint())
}

func TestBristolInteractorTargets(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	op.On("TargetsOutput", bg, "tableau", 2).Return("targets_output")
	assert.Equal(t, "targets_output", bi.Targets("tableau", 2))
}

func TestBristolInteractorAutoComplete(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("AutoComplete").Return(nil)
	op.On("Output", bg, nil).Return("ac")
	assert.Equal(t, "ac", bi.AutoComplete())
}

func TestBristolInteractorActionLog(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	op.On("ActionLogOutput", mock.Anything).Return("log")
	assert.Equal(t, "log", bi.ActionLog())
}

func TestBristolInteractorUndo(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("Undo").Return(nil)
	op.On("Output", bg, nil).Return("undo")
	assert.Equal(t, "undo", bi.Undo())
}

func TestBristolInteractorUndoN(t *testing.T) {
	bg := newMockBristolGame()
	op := newMockBristolPresenter()
	bi := NewBristolInteractor(bg, op)
	bg.On("UndoN", 3).Return(nil)
	op.On("Output", bg, nil).Return("undon")
	assert.Equal(t, "undon", bi.UndoN(3))
}

func TestRestoreBristolInteractor(t *testing.T) {
	src := domain.NewDefaultBristol()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	bi, err := RestoreBristolInteractor(data, newMockBristolPresenter())
	assert.NoError(t, err)
	assert.NotNil(t, bi)

	_, err = RestoreBristolInteractor([]byte("not json"), newMockBristolPresenter())
	assert.Error(t, err)
}

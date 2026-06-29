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

func newMockAgnesGame() *interfaces.MockAgnesGame {
	return new(interfaces.MockAgnesGame)
}

func newMockAgnesPresenter() *presenter.MockAgnesPresenter {
	return new(presenter.MockAgnesPresenter)
}

func TestNewAgnesInteractor(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	assert.NotNil(t, ai)
}

func TestNewAgnesInteractorPanicsOnNil(t *testing.T) {
	cp := newMockAgnesPresenter()
	assert.Panics(t, func() { NewAgnesInteractor(nil, cp) })
	cg := newMockAgnesGame()
	assert.Panics(t, func() { NewAgnesInteractor(cg, nil) })
}

func TestAgnesInteractorReset(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	cg.On("Reset").Return()
	cp.On("Output", cg, nil).Return("reset_output")
	assert.Equal(t, "reset_output", ai.Reset())
}

func TestAgnesInteractorDealStock(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockAgnesGame()
		cp := newMockAgnesPresenter()
		ai := NewAgnesInteractor(cg, cp)
		cg.On("DealStock").Return(nil)
		cp.On("Output", cg, nil).Return("deal_output")
		assert.Equal(t, "deal_output", ai.DealStock())
	})
	t.Run("error", func(t *testing.T) {
		cg := newMockAgnesGame()
		cp := newMockAgnesPresenter()
		ai := NewAgnesInteractor(cg, cp)
		err := errors.New("stock empty")
		cg.On("DealStock").Return(err)
		cp.On("Output", cg, err).Return("error_output")
		assert.Equal(t, "error_output", ai.DealStock())
	})
}

func TestAgnesInteractorMoveTableauToTableau(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	cg.On("MoveTableauToTableau", 0, -1, 2).Return(nil)
	cp.On("Output", cg, nil).Return("ok")
	assert.Equal(t, "ok", ai.MoveTableauToTableau(0, -1, 2))
}

func TestAgnesInteractorMoveTableauToFoundation(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	cg.On("MoveTableauToFoundation", 1).Return(nil)
	cp.On("Output", cg, nil).Return("ok")
	assert.Equal(t, "ok", ai.MoveTableauToFoundation(1))
}

func TestAgnesInteractorGiveUp(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	cg.On("GiveUp").Return()
	cp.On("Output", cg, nil).Return("giveup")
	assert.Equal(t, "giveup", ai.GiveUp())
}

func TestAgnesInteractorHint(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	cp.On("HintOutput", mock.Anything).Return("hint")
	assert.Equal(t, "hint", ai.Hint())
}

func TestAgnesInteractorActionLog(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	cp.On("ActionLogOutput", mock.Anything).Return("log")
	assert.Equal(t, "log", ai.ActionLog())
}

func TestAgnesInteractorUndo(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	cg.On("Undo").Return(nil)
	cp.On("Output", cg, nil).Return("undo")
	assert.Equal(t, "undo", ai.Undo())
}

func TestAgnesInteractorUndoN(t *testing.T) {
	cg := newMockAgnesGame()
	cp := newMockAgnesPresenter()
	ai := NewAgnesInteractor(cg, cp)
	cg.On("UndoN", 3).Return(nil)
	cp.On("Output", cg, nil).Return("undon")
	assert.Equal(t, "undon", ai.UndoN(3))
}

func TestRestoreAgnesInteractor(t *testing.T) {
	src := domain.NewDefaultAgnes()
	src.Reset()
	data, err := src.MarshalJSON()
	assert.NoError(t, err)
	cp := newMockAgnesPresenter()
	ai, err := RestoreAgnesInteractor(data, cp)
	assert.NoError(t, err)
	assert.NotNil(t, ai)
}

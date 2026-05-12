package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockSpideretteGame() *interfaces.MockSpideretteGame {
	return new(interfaces.MockSpideretteGame)
}

func newMockSpiderettePresenter() *presenter.MockSpiderettePresenter {
	return new(presenter.MockSpiderettePresenter)
}

func TestNewSpideretteInteractor(t *testing.T) {
	sg := newMockSpideretteGame()
	sp := newMockSpiderettePresenter()
	si := NewSpideretteInteractor(sg, sp)
	assert.NotNil(t, si)
}

func TestNewSpideretteInteractorPanicsOnNil(t *testing.T) {
	sp := newMockSpiderettePresenter()
	assert.Panics(t, func() { NewSpideretteInteractor(nil, sp) })
	sg := newMockSpideretteGame()
	assert.Panics(t, func() { NewSpideretteInteractor(sg, nil) })
}

func TestSpideretteInteractorReset(t *testing.T) {
	sg := newMockSpideretteGame()
	sp := newMockSpiderettePresenter()
	si := NewSpideretteInteractor(sg, sp)

	sg.On("Reset").Return()
	sp.On("Output", sg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", si.Reset())
	sg.AssertCalled(t, "Reset")
}

func TestSpideretteInteractorDeal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSpideretteGame()
		sp := newMockSpiderettePresenter()
		si := NewSpideretteInteractor(sg, sp)
		sg.On("Deal").Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.Deal())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSpideretteGame()
		sp := newMockSpiderettePresenter()
		si := NewSpideretteInteractor(sg, sp)
		err := errors.New("no stock")
		sg.On("Deal").Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.Deal())
	})
}

func TestSpideretteInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSpideretteGame()
		sp := newMockSpiderettePresenter()
		si := NewSpideretteInteractor(sg, sp)
		sg.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.MoveTableauToTableau(0, 2, 3))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSpideretteGame()
		sp := newMockSpiderettePresenter()
		si := NewSpideretteInteractor(sg, sp)
		err := errors.New("invalid")
		sg.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.MoveTableauToTableau(0, 2, 3))
	})
}

func TestSpideretteInteractorGiveUp(t *testing.T) {
	sg := newMockSpideretteGame()
	sp := newMockSpiderettePresenter()
	si := NewSpideretteInteractor(sg, sp)
	sg.On("GiveUp").Return()
	sp.On("Output", sg, nil).Return("ok")
	assert.Equal(t, "ok", si.GiveUp())
}

func TestSpideretteInteractorHint(t *testing.T) {
	sg := newMockSpideretteGame()
	sp := newMockSpiderettePresenter()
	si := NewSpideretteInteractor(sg, sp)
	sp.On("HintOutput", mock.Anything).Return("hint")
	assert.Equal(t, "hint", si.Hint())
}

func TestSpideretteInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSpideretteGame()
		sp := newMockSpiderettePresenter()
		si := NewSpideretteInteractor(sg, sp)
		sg.On("AutoComplete").Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.AutoComplete())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSpideretteGame()
		sp := newMockSpiderettePresenter()
		si := NewSpideretteInteractor(sg, sp)
		err := errors.New("nope")
		sg.On("AutoComplete").Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.AutoComplete())
	})
}

func TestSpideretteInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSpideretteGame()
		sp := newMockSpiderettePresenter()
		si := NewSpideretteInteractor(sg, sp)
		sg.On("Undo").Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.Undo())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSpideretteGame()
		sp := newMockSpiderettePresenter()
		si := NewSpideretteInteractor(sg, sp)
		err := errors.New("no history")
		sg.On("Undo").Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.Undo())
	})
}

func TestSpideretteInteractorUndoN(t *testing.T) {
	sg := newMockSpideretteGame()
	sp := newMockSpiderettePresenter()
	si := NewSpideretteInteractor(sg, sp)
	sg.On("UndoN", 3).Return(nil)
	sp.On("Output", sg, nil).Return("ok")
	assert.Equal(t, "ok", si.UndoN(3))
}

func TestSpideretteInteractorActionLog(t *testing.T) {
	sg := newMockSpideretteGame()
	sp := newMockSpiderettePresenter()
	si := NewSpideretteInteractor(sg, sp)
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	assert.Equal(t, "log", si.ActionLog())
}

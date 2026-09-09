package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockWillOTheWispGame() *interfaces.MockWillOTheWispGame {
	return new(interfaces.MockWillOTheWispGame)
}

func newMockWillOTheWispPresenter() *presenter.MockWillOTheWispPresenter {
	return new(presenter.MockWillOTheWispPresenter)
}

func TestNewWillOTheWispInteractor(t *testing.T) {
	sg := newMockWillOTheWispGame()
	sp := newMockWillOTheWispPresenter()
	si := NewWillOTheWispInteractor(sg, sp)
	assert.NotNil(t, si)
}

func TestNewWillOTheWispInteractorPanicsOnNil(t *testing.T) {
	sp := newMockWillOTheWispPresenter()
	assert.Panics(t, func() { NewWillOTheWispInteractor(nil, sp) })
	sg := newMockWillOTheWispGame()
	assert.Panics(t, func() { NewWillOTheWispInteractor(sg, nil) })
}

func TestWillOTheWispInteractorReset(t *testing.T) {
	sg := newMockWillOTheWispGame()
	sp := newMockWillOTheWispPresenter()
	si := NewWillOTheWispInteractor(sg, sp)

	sg.On("Reset").Return()
	sp.On("Output", sg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", si.Reset())
	sg.AssertCalled(t, "Reset")
}

func TestWillOTheWispInteractorDeal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockWillOTheWispGame()
		sp := newMockWillOTheWispPresenter()
		si := NewWillOTheWispInteractor(sg, sp)
		sg.On("Deal").Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.Deal())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockWillOTheWispGame()
		sp := newMockWillOTheWispPresenter()
		si := NewWillOTheWispInteractor(sg, sp)
		err := errors.New("no stock")
		sg.On("Deal").Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.Deal())
	})
}

func TestWillOTheWispInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockWillOTheWispGame()
		sp := newMockWillOTheWispPresenter()
		si := NewWillOTheWispInteractor(sg, sp)
		sg.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.MoveTableauToTableau(0, 2, 3))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockWillOTheWispGame()
		sp := newMockWillOTheWispPresenter()
		si := NewWillOTheWispInteractor(sg, sp)
		err := errors.New("invalid")
		sg.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.MoveTableauToTableau(0, 2, 3))
	})
}

func TestWillOTheWispInteractorGiveUp(t *testing.T) {
	sg := newMockWillOTheWispGame()
	sp := newMockWillOTheWispPresenter()
	si := NewWillOTheWispInteractor(sg, sp)
	sg.On("GiveUp").Return()
	sp.On("Output", sg, nil).Return("ok")
	assert.Equal(t, "ok", si.GiveUp())
}

func TestWillOTheWispInteractorHint(t *testing.T) {
	sg := newMockWillOTheWispGame()
	sp := newMockWillOTheWispPresenter()
	si := NewWillOTheWispInteractor(sg, sp)
	sp.On("HintOutput", mock.Anything).Return("hint")
	assert.Equal(t, "hint", si.Hint())
}

func TestWillOTheWispInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockWillOTheWispGame()
		sp := newMockWillOTheWispPresenter()
		si := NewWillOTheWispInteractor(sg, sp)
		sg.On("AutoComplete").Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.AutoComplete())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockWillOTheWispGame()
		sp := newMockWillOTheWispPresenter()
		si := NewWillOTheWispInteractor(sg, sp)
		err := errors.New("nope")
		sg.On("AutoComplete").Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.AutoComplete())
	})
}

func TestWillOTheWispInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockWillOTheWispGame()
		sp := newMockWillOTheWispPresenter()
		si := NewWillOTheWispInteractor(sg, sp)
		sg.On("Undo").Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.Undo())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockWillOTheWispGame()
		sp := newMockWillOTheWispPresenter()
		si := NewWillOTheWispInteractor(sg, sp)
		err := errors.New("no history")
		sg.On("Undo").Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.Undo())
	})
}

func TestWillOTheWispInteractorUndoN(t *testing.T) {
	sg := newMockWillOTheWispGame()
	sp := newMockWillOTheWispPresenter()
	si := NewWillOTheWispInteractor(sg, sp)
	sg.On("UndoN", 3).Return(nil)
	sp.On("Output", sg, nil).Return("ok")
	assert.Equal(t, "ok", si.UndoN(3))
}

func TestWillOTheWispInteractorActionLog(t *testing.T) {
	sg := newMockWillOTheWispGame()
	sp := newMockWillOTheWispPresenter()
	si := NewWillOTheWispInteractor(sg, sp)
	sp.On("ActionLogOutput", mock.Anything).Return("log")
	assert.Equal(t, "log", si.ActionLog())
}

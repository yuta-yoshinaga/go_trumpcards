package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockCrescentGame() *interfaces.MockCrescentGame {
	return new(interfaces.MockCrescentGame)
}

func newMockCrescentPresenter() *presenter.MockCrescentPresenter {
	return new(presenter.MockCrescentPresenter)
}

func TestNewCrescentInteractor(t *testing.T) {
	cg := newMockCrescentGame()
	cp := newMockCrescentPresenter()
	ci := NewCrescentInteractor(cg, cp)
	assert.NotNil(t, ci)
}

func TestNewCrescentInteractorPanicsOnNil(t *testing.T) {
	cp := newMockCrescentPresenter()
	assert.Panics(t, func() { NewCrescentInteractor(nil, cp) })
	cg := newMockCrescentGame()
	assert.Panics(t, func() { NewCrescentInteractor(cg, nil) })
}

func TestCrescentInteractorReset(t *testing.T) {
	cg := newMockCrescentGame()
	cp := newMockCrescentPresenter()
	ci := NewCrescentInteractor(cg, cp)
	cg.On("Reset").Return()
	cp.On("Output", cg, nil).Return("reset_output")
	assert.Equal(t, "reset_output", ci.Reset())
	cg.AssertCalled(t, "Reset")
}

func TestCrescentInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		cg.On("MoveTableauToTableau", 0, 5).Return(nil)
		cp.On("Output", cg, nil).Return("move_output")
		assert.Equal(t, "move_output", ci.MoveTableauToTableau(0, 5))
	})
	t.Run("error", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		err := errors.New("invalid")
		cg.On("MoveTableauToTableau", 0, 5).Return(err)
		cp.On("Output", cg, err).Return("error_output")
		assert.Equal(t, "error_output", ci.MoveTableauToTableau(0, 5))
	})
}

func TestCrescentInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		cg.On("MoveTableauToFoundation", 2, 4).Return(nil)
		cp.On("Output", cg, nil).Return("move_output")
		assert.Equal(t, "move_output", ci.MoveTableauToFoundation(2, 4))
	})
	t.Run("error", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		err := errors.New("invalid foundation index")
		cg.On("MoveTableauToFoundation", 2, 4).Return(err)
		cp.On("Output", cg, err).Return("error_output")
		assert.Equal(t, "error_output", ci.MoveTableauToFoundation(2, 4))
	})
}

func TestCrescentInteractorRedeal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		cg.On("Redeal").Return(nil)
		cp.On("Output", cg, nil).Return("redeal_output")
		assert.Equal(t, "redeal_output", ci.Redeal())
	})
	t.Run("error", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		err := errors.New("no redeals remaining")
		cg.On("Redeal").Return(err)
		cp.On("Output", cg, err).Return("error_output")
		assert.Equal(t, "error_output", ci.Redeal())
	})
}

func TestCrescentInteractorGiveUp(t *testing.T) {
	cg := newMockCrescentGame()
	cp := newMockCrescentPresenter()
	ci := NewCrescentInteractor(cg, cp)
	cg.On("GiveUp").Return()
	cp.On("Output", cg, nil).Return("giveup_output")
	assert.Equal(t, "giveup_output", ci.GiveUp())
}

func TestCrescentInteractorHint(t *testing.T) {
	cg := newMockCrescentGame()
	cp := newMockCrescentPresenter()
	ci := NewCrescentInteractor(cg, cp)
	cp.On("HintOutput", mock.Anything).Return("hint_output")
	assert.Equal(t, "hint_output", ci.Hint())
}

func TestCrescentInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		cg.On("AutoComplete").Return(nil)
		cp.On("Output", cg, nil).Return("ac_output")
		assert.Equal(t, "ac_output", ci.AutoComplete())
	})
	t.Run("error", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		err := errors.New("not playing")
		cg.On("AutoComplete").Return(err)
		cp.On("Output", cg, err).Return("error_output")
		assert.Equal(t, "error_output", ci.AutoComplete())
	})
}

func TestCrescentInteractorActionLog(t *testing.T) {
	cg := newMockCrescentGame()
	cp := newMockCrescentPresenter()
	ci := NewCrescentInteractor(cg, cp)
	cp.On("ActionLogOutput", mock.Anything).Return("log_output")
	assert.Equal(t, "log_output", ci.ActionLog())
}

func TestCrescentInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		cg.On("Undo").Return(nil)
		cp.On("Output", cg, nil).Return("undo_output")
		assert.Equal(t, "undo_output", ci.Undo())
	})
	t.Run("error", func(t *testing.T) {
		cg := newMockCrescentGame()
		cp := newMockCrescentPresenter()
		ci := NewCrescentInteractor(cg, cp)
		err := errors.New("nothing to undo")
		cg.On("Undo").Return(err)
		cp.On("Output", cg, err).Return("error_output")
		assert.Equal(t, "error_output", ci.Undo())
	})
}

func TestCrescentInteractorUndoN(t *testing.T) {
	cg := newMockCrescentGame()
	cp := newMockCrescentPresenter()
	ci := NewCrescentInteractor(cg, cp)
	cg.On("UndoN", 2).Return(nil)
	cp.On("Output", cg, nil).Return("undon_output")
	assert.Equal(t, "undon_output", ci.UndoN(2))
}

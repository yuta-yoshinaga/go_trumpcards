//go:build test

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockPenguinGame() *interfaces.MockPenguinGame {
	return new(interfaces.MockPenguinGame)
}

func newMockPenguinPresenter() *presenter.MockPenguinPresenter {
	return new(presenter.MockPenguinPresenter)
}

func TestNewPenguinInteractor(t *testing.T) {
	g := newMockPenguinGame()
	p := newMockPenguinPresenter()
	pi := NewPenguinInteractor(g, p)
	assert.NotNil(t, pi)
}

func TestNewPenguinInteractorPanicsOnNil(t *testing.T) {
	p := newMockPenguinPresenter()
	assert.Panics(t, func() { NewPenguinInteractor(nil, p) })
	g := newMockPenguinGame()
	assert.Panics(t, func() { NewPenguinInteractor(g, nil) })
}

func TestPenguinInteractorReset(t *testing.T) {
	g := newMockPenguinGame()
	p := newMockPenguinPresenter()
	pi := NewPenguinInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	result := pi.Reset()
	assert.Equal(t, "reset_output", result)
	g.AssertCalled(t, "Reset")
}

func TestPenguinInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		g.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", pi.MoveTableauToTableau(0, 2, 3))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		err := errors.New("invalid")
		g.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", pi.MoveTableauToTableau(0, 2, 3))
	})
}

func TestPenguinInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		g.On("MoveTableauToFoundation", 2).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", pi.MoveTableauToFoundation(2))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		err := errors.New("invalid")
		g.On("MoveTableauToFoundation", 2).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", pi.MoveTableauToFoundation(2))
	})
}

func TestPenguinInteractorMoveTableauToFreeCell(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		g.On("MoveTableauToFreeCell", 0, 5).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", pi.MoveTableauToFreeCell(0, 5))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		err := errors.New("occupied")
		g.On("MoveTableauToFreeCell", 0, 5).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", pi.MoveTableauToFreeCell(0, 5))
	})
}

func TestPenguinInteractorMoveFreeCellToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		g.On("MoveFreeCellToTableau", 1, 3).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", pi.MoveFreeCellToTableau(1, 3))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		err := errors.New("empty")
		g.On("MoveFreeCellToTableau", 1, 3).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", pi.MoveFreeCellToTableau(1, 3))
	})
}

func TestPenguinInteractorMoveFreeCellToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		g.On("MoveFreeCellToFoundation", 0).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", pi.MoveFreeCellToFoundation(0))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		err := errors.New("invalid")
		g.On("MoveFreeCellToFoundation", 0).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", pi.MoveFreeCellToFoundation(0))
	})
}

func TestPenguinInteractorGiveUp(t *testing.T) {
	g := newMockPenguinGame()
	p := newMockPenguinPresenter()
	pi := NewPenguinInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", pi.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestPenguinInteractorHint(t *testing.T) {
	g := newMockPenguinGame()
	p := newMockPenguinPresenter()
	pi := NewPenguinInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")

	assert.Equal(t, "hint_output", pi.Hint())
}

func TestPenguinInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		g.On("AutoComplete").Return(nil)
		p.On("Output", g, nil).Return("ac_output")

		assert.Equal(t, "ac_output", pi.AutoComplete())
	})
	t.Run("error", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		err := errors.New("not playing")
		g.On("AutoComplete").Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", pi.AutoComplete())
	})
}

func TestPenguinInteractorActionLog(t *testing.T) {
	g := newMockPenguinGame()
	p := newMockPenguinPresenter()
	pi := NewPenguinInteractor(g, p)

	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "log_output", pi.ActionLog())
}

func TestPenguinInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		g.On("Undo").Return(nil)
		p.On("Output", g, nil).Return("undo_output")

		assert.Equal(t, "undo_output", pi.Undo())
	})
	t.Run("error", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		err := errors.New("no history")
		g.On("Undo").Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", pi.Undo())
	})
}

func TestPenguinInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		g.On("UndoN", 2).Return(nil)
		p.On("Output", g, nil).Return("undo_output")

		assert.Equal(t, "undo_output", pi.UndoN(2))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockPenguinGame()
		p := newMockPenguinPresenter()
		pi := NewPenguinInteractor(g, p)

		err := errors.New("no history")
		g.On("UndoN", 2).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", pi.UndoN(2))
	})
}

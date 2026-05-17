package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockEightOffGame() *interfaces.MockEightOffGame {
	return new(interfaces.MockEightOffGame)
}

func newMockEightOffPresenter() *presenter.MockEightOffPresenter {
	return new(presenter.MockEightOffPresenter)
}

func TestNewEightOffInteractor(t *testing.T) {
	g := newMockEightOffGame()
	p := newMockEightOffPresenter()
	ei := NewEightOffInteractor(g, p)
	assert.NotNil(t, ei)
}

func TestNewEightOffInteractorPanicsOnNil(t *testing.T) {
	p := newMockEightOffPresenter()
	assert.Panics(t, func() { NewEightOffInteractor(nil, p) })
	g := newMockEightOffGame()
	assert.Panics(t, func() { NewEightOffInteractor(g, nil) })
}

func TestEightOffInteractorReset(t *testing.T) {
	g := newMockEightOffGame()
	p := newMockEightOffPresenter()
	ei := NewEightOffInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	result := ei.Reset()
	assert.Equal(t, "reset_output", result)
	g.AssertCalled(t, "Reset")
}

func TestEightOffInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		g.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", ei.MoveTableauToTableau(0, 2, 3))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		err := errors.New("invalid")
		g.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", ei.MoveTableauToTableau(0, 2, 3))
	})
}

func TestEightOffInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		g.On("MoveTableauToFoundation", 2).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", ei.MoveTableauToFoundation(2))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		err := errors.New("invalid")
		g.On("MoveTableauToFoundation", 2).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", ei.MoveTableauToFoundation(2))
	})
}

func TestEightOffInteractorMoveTableauToFreeCell(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		g.On("MoveTableauToFreeCell", 0, 5).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", ei.MoveTableauToFreeCell(0, 5))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		err := errors.New("occupied")
		g.On("MoveTableauToFreeCell", 0, 5).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", ei.MoveTableauToFreeCell(0, 5))
	})
}

func TestEightOffInteractorMoveFreeCellToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		g.On("MoveFreeCellToTableau", 1, 3).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", ei.MoveFreeCellToTableau(1, 3))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		err := errors.New("empty")
		g.On("MoveFreeCellToTableau", 1, 3).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", ei.MoveFreeCellToTableau(1, 3))
	})
}

func TestEightOffInteractorMoveFreeCellToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		g.On("MoveFreeCellToFoundation", 0).Return(nil)
		p.On("Output", g, nil).Return("move_output")

		assert.Equal(t, "move_output", ei.MoveFreeCellToFoundation(0))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		err := errors.New("invalid")
		g.On("MoveFreeCellToFoundation", 0).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", ei.MoveFreeCellToFoundation(0))
	})
}

func TestEightOffInteractorGiveUp(t *testing.T) {
	g := newMockEightOffGame()
	p := newMockEightOffPresenter()
	ei := NewEightOffInteractor(g, p)

	g.On("GiveUp").Return()
	p.On("Output", g, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", ei.GiveUp())
	g.AssertCalled(t, "GiveUp")
}

func TestEightOffInteractorHint(t *testing.T) {
	g := newMockEightOffGame()
	p := newMockEightOffPresenter()
	ei := NewEightOffInteractor(g, p)

	p.On("HintOutput", mock.Anything).Return("hint_output")

	assert.Equal(t, "hint_output", ei.Hint())
}

func TestEightOffInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		g.On("AutoComplete").Return(nil)
		p.On("Output", g, nil).Return("ac_output")

		assert.Equal(t, "ac_output", ei.AutoComplete())
	})
	t.Run("error", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		err := errors.New("not playing")
		g.On("AutoComplete").Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", ei.AutoComplete())
	})
}

func TestEightOffInteractorActionLog(t *testing.T) {
	g := newMockEightOffGame()
	p := newMockEightOffPresenter()
	ei := NewEightOffInteractor(g, p)

	p.On("ActionLogOutput", mock.Anything).Return("log_output")

	assert.Equal(t, "log_output", ei.ActionLog())
}

func TestEightOffInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		g.On("Undo").Return(nil)
		p.On("Output", g, nil).Return("undo_output")

		assert.Equal(t, "undo_output", ei.Undo())
	})
	t.Run("error", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		err := errors.New("no history")
		g.On("Undo").Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", ei.Undo())
	})
}

func TestEightOffInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		g.On("UndoN", 2).Return(nil)
		p.On("Output", g, nil).Return("undo_output")

		assert.Equal(t, "undo_output", ei.UndoN(2))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockEightOffGame()
		p := newMockEightOffPresenter()
		ei := NewEightOffInteractor(g, p)

		err := errors.New("no history")
		g.On("UndoN", 2).Return(err)
		p.On("Output", g, err).Return("error_output")

		assert.Equal(t, "error_output", ei.UndoN(2))
	})
}

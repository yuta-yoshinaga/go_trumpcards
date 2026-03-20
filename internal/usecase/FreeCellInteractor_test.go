package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockFreeCellGame() *interfaces.MockFreeCellGame {
	return new(interfaces.MockFreeCellGame)
}

func newMockFreeCellPresenter() *presenter.MockFreeCellPresenter {
	return new(presenter.MockFreeCellPresenter)
}

func TestNewFreeCellInteractor(t *testing.T) {
	fg := newMockFreeCellGame()
	fp := newMockFreeCellPresenter()
	fi := NewFreeCellInteractor(fg, fp)
	assert.NotNil(t, fi)
}

func TestNewFreeCellInteractorPanicsOnNil(t *testing.T) {
	fp := newMockFreeCellPresenter()
	assert.Panics(t, func() { NewFreeCellInteractor(nil, fp) })
	fg := newMockFreeCellGame()
	assert.Panics(t, func() { NewFreeCellInteractor(fg, nil) })
}

func TestFreeCellInteractorReset(t *testing.T) {
	fg := newMockFreeCellGame()
	fp := newMockFreeCellPresenter()
	fi := NewFreeCellInteractor(fg, fp)

	fg.On("Reset").Return()
	fp.On("Output", fg, nil).Return("reset_output")

	result := fi.Reset()
	assert.Equal(t, "reset_output", result)
	fg.AssertCalled(t, "Reset")
}

func TestFreeCellInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		fg.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		err := errors.New("invalid")
		fg.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "error_output", result)
	})
}

func TestFreeCellInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		fg.On("MoveTableauToFoundation", 2).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		err := errors.New("invalid")
		fg.On("MoveTableauToFoundation", 2).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestFreeCellInteractorMoveTableauToFreeCell(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		fg.On("MoveTableauToFreeCell", 0, 1).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveTableauToFreeCell(0, 1)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		err := errors.New("occupied")
		fg.On("MoveTableauToFreeCell", 0, 1).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveTableauToFreeCell(0, 1)
		assert.Equal(t, "error_output", result)
	})
}

func TestFreeCellInteractorMoveFreeCellToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		fg.On("MoveFreeCellToTableau", 1, 3).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveFreeCellToTableau(1, 3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		err := errors.New("empty")
		fg.On("MoveFreeCellToTableau", 1, 3).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveFreeCellToTableau(1, 3)
		assert.Equal(t, "error_output", result)
	})
}

func TestFreeCellInteractorMoveFreeCellToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		fg.On("MoveFreeCellToFoundation", 0).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveFreeCellToFoundation(0)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		err := errors.New("invalid")
		fg.On("MoveFreeCellToFoundation", 0).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveFreeCellToFoundation(0)
		assert.Equal(t, "error_output", result)
	})
}

func TestFreeCellInteractorGiveUp(t *testing.T) {
	fg := newMockFreeCellGame()
	fp := newMockFreeCellPresenter()
	fi := NewFreeCellInteractor(fg, fp)

	fg.On("GiveUp").Return()
	fp.On("Output", fg, nil).Return("giveup_output")

	result := fi.GiveUp()
	assert.Equal(t, "giveup_output", result)
	fg.AssertCalled(t, "GiveUp")
}

func TestFreeCellInteractorHint(t *testing.T) {
	fg := newMockFreeCellGame()
	fp := newMockFreeCellPresenter()
	fi := NewFreeCellInteractor(fg, fp)

	fp.On("HintOutput", mock.Anything).Return("hint_output")

	result := fi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestFreeCellInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		fg.On("AutoComplete").Return(nil)
		fp.On("Output", fg, nil).Return("ac_output")

		result := fi.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		err := errors.New("not playing")
		fg.On("AutoComplete").Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestFreeCellInteractorActionLog(t *testing.T) {
	fg := newMockFreeCellGame()
	fp := newMockFreeCellPresenter()
	fi := NewFreeCellInteractor(fg, fp)

	fp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := fi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestFreeCellInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		fg.On("Undo").Return(nil)
		fp.On("Output", fg, nil).Return("undo_output")

		result := fi.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockFreeCellGame()
		fp := newMockFreeCellPresenter()
		fi := NewFreeCellInteractor(fg, fp)

		err := errors.New("nothing to undo")
		fg.On("Undo").Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.Undo()
		assert.Equal(t, "error_output", result)
	})
}

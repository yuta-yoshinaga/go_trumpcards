package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockRankAndFileGame() *interfaces.MockRankAndFileGame {
	return new(interfaces.MockRankAndFileGame)
}

func newMockRankAndFilePresenter() *presenter.MockRankAndFilePresenter {
	return new(presenter.MockRankAndFilePresenter)
}

func TestNewRankAndFileInteractor(t *testing.T) {
	fg := newMockRankAndFileGame()
	fp := newMockRankAndFilePresenter()
	fi := NewRankAndFileInteractor(fg, fp)
	assert.NotNil(t, fi)
}

func TestNewRankAndFileInteractorPanicsOnNil(t *testing.T) {
	fp := newMockRankAndFilePresenter()
	assert.Panics(t, func() { NewRankAndFileInteractor(nil, fp) })
	fg := newMockRankAndFileGame()
	assert.Panics(t, func() { NewRankAndFileInteractor(fg, nil) })
}

func TestRankAndFileInteractorReset(t *testing.T) {
	fg := newMockRankAndFileGame()
	fp := newMockRankAndFilePresenter()
	fi := NewRankAndFileInteractor(fg, fp)

	fg.On("Reset").Return()
	fp.On("Output", fg, nil).Return("reset_output")

	result := fi.Reset()
	assert.Equal(t, "reset_output", result)
	fg.AssertCalled(t, "Reset")
}

func TestRankAndFileInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		fg.On("Draw").Return(nil)
		fp.On("Output", fg, nil).Return("draw_output")

		result := fi.Draw()
		assert.Equal(t, "draw_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		err := errors.New("no cards")
		fg.On("Draw").Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.Draw()
		assert.Equal(t, "error_output", result)
	})
}

func TestRankAndFileInteractorMoveWasteToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		fg.On("MoveWasteToTableau", 3).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveWasteToTableau(3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		err := errors.New("cannot place")
		fg.On("MoveWasteToTableau", 3).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveWasteToTableau(3)
		assert.Equal(t, "error_output", result)
	})
}

func TestRankAndFileInteractorMoveWasteToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		fg.On("MoveWasteToFoundation").Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveWasteToFoundation()
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		err := errors.New("cannot place")
		fg.On("MoveWasteToFoundation").Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveWasteToFoundation()
		assert.Equal(t, "error_output", result)
	})
}

func TestRankAndFileInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		fg.On("MoveTableauToTableau", 0, 3, 5).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		err := errors.New("invalid")
		fg.On("MoveTableauToTableau", 0, 3, 5).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "error_output", result)
	})
}

func TestRankAndFileInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		fg.On("MoveTableauToFoundation", 2).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		err := errors.New("invalid")
		fg.On("MoveTableauToFoundation", 2).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestRankAndFileInteractorGiveUp(t *testing.T) {
	fg := newMockRankAndFileGame()
	fp := newMockRankAndFilePresenter()
	fi := NewRankAndFileInteractor(fg, fp)

	fg.On("GiveUp").Return()
	fp.On("Output", fg, nil).Return("giveup_output")

	result := fi.GiveUp()
	assert.Equal(t, "giveup_output", result)
	fg.AssertCalled(t, "GiveUp")
}

func TestRankAndFileInteractorHint(t *testing.T) {
	fg := newMockRankAndFileGame()
	fp := newMockRankAndFilePresenter()
	fi := NewRankAndFileInteractor(fg, fp)

	fp.On("HintOutput", mock.Anything).Return("hint_output")

	result := fi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestRankAndFileInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		fg.On("AutoComplete").Return(nil)
		fp.On("Output", fg, nil).Return("ac_output")

		result := fi.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		err := errors.New("not all face up")
		fg.On("AutoComplete").Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestRankAndFileInteractorActionLog(t *testing.T) {
	fg := newMockRankAndFileGame()
	fp := newMockRankAndFilePresenter()
	fi := NewRankAndFileInteractor(fg, fp)

	fp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := fi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestRankAndFileInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		fg.On("Undo").Return(nil)
		fp.On("Output", fg, nil).Return("undo_output")

		result := fi.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		err := errors.New("nothing to undo")
		fg.On("Undo").Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.Undo()
		assert.Equal(t, "error_output", result)
	})
}

func TestRankAndFileInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		fg.On("UndoN", 3).Return(nil)
		fp.On("Output", fg, nil).Return("undon_output")

		result := fi.UndoN(3)
		assert.Equal(t, "undon_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockRankAndFileGame()
		fp := newMockRankAndFilePresenter()
		fi := NewRankAndFileInteractor(fg, fp)

		err := errors.New("undo failed")
		fg.On("UndoN", 3).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.UndoN(3)
		assert.Equal(t, "error_output", result)
	})
}

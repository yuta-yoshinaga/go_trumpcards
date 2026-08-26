package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockStalactitesGame() *interfaces.MockStalactitesGame {
	return new(interfaces.MockStalactitesGame)
}

func newMockStalactitesPresenter() *presenter.MockStalactitesPresenter {
	return new(presenter.MockStalactitesPresenter)
}

func TestNewStalactitesInteractor(t *testing.T) {
	fg := newMockStalactitesGame()
	fp := newMockStalactitesPresenter()
	fi := NewStalactitesInteractor(fg, fp)
	assert.NotNil(t, fi)
}

func TestNewStalactitesInteractorPanicsOnNil(t *testing.T) {
	fp := newMockStalactitesPresenter()
	assert.Panics(t, func() { NewStalactitesInteractor(nil, fp) })
	fg := newMockStalactitesGame()
	assert.Panics(t, func() { NewStalactitesInteractor(fg, nil) })
}

func TestStalactitesInteractorReset(t *testing.T) {
	fg := newMockStalactitesGame()
	fp := newMockStalactitesPresenter()
	fi := NewStalactitesInteractor(fg, fp)

	fg.On("Reset").Return()
	fp.On("Output", fg, nil).Return("reset_output")

	result := fi.Reset()
	assert.Equal(t, "reset_output", result)
	fg.AssertCalled(t, "Reset")
}

func TestStalactitesInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		fg.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		err := errors.New("invalid")
		fg.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "error_output", result)
	})
}

func TestStalactitesInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		fg.On("MoveTableauToFoundation", 2).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		err := errors.New("invalid")
		fg.On("MoveTableauToFoundation", 2).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestStalactitesInteractorMoveTableauToStalactites(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		fg.On("MoveTableauToStalactites", 0, 1).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveTableauToStalactites(0, 1)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		err := errors.New("occupied")
		fg.On("MoveTableauToStalactites", 0, 1).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveTableauToStalactites(0, 1)
		assert.Equal(t, "error_output", result)
	})
}

func TestStalactitesInteractorMoveStalactitesToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		fg.On("MoveStalactitesToTableau", 1, 3).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveStalactitesToTableau(1, 3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		err := errors.New("empty")
		fg.On("MoveStalactitesToTableau", 1, 3).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveStalactitesToTableau(1, 3)
		assert.Equal(t, "error_output", result)
	})
}

func TestStalactitesInteractorMoveStalactitesToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		fg.On("MoveStalactitesToFoundation", 0).Return(nil)
		fp.On("Output", fg, nil).Return("move_output")

		result := fi.MoveStalactitesToFoundation(0)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		err := errors.New("invalid")
		fg.On("MoveStalactitesToFoundation", 0).Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.MoveStalactitesToFoundation(0)
		assert.Equal(t, "error_output", result)
	})
}

func TestStalactitesInteractorGiveUp(t *testing.T) {
	fg := newMockStalactitesGame()
	fp := newMockStalactitesPresenter()
	fi := NewStalactitesInteractor(fg, fp)

	fg.On("GiveUp").Return()
	fp.On("Output", fg, nil).Return("giveup_output")

	result := fi.GiveUp()
	assert.Equal(t, "giveup_output", result)
	fg.AssertCalled(t, "GiveUp")
}

func TestStalactitesInteractorHint(t *testing.T) {
	fg := newMockStalactitesGame()
	fp := newMockStalactitesPresenter()
	fi := NewStalactitesInteractor(fg, fp)

	fp.On("HintOutput", mock.Anything).Return("hint_output")

	result := fi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestStalactitesInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		fg.On("AutoComplete").Return(nil)
		fp.On("Output", fg, nil).Return("ac_output")

		result := fi.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		err := errors.New("not playing")
		fg.On("AutoComplete").Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestStalactitesInteractorActionLog(t *testing.T) {
	fg := newMockStalactitesGame()
	fp := newMockStalactitesPresenter()
	fi := NewStalactitesInteractor(fg, fp)

	fp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := fi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestStalactitesInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		fg.On("Undo").Return(nil)
		fp.On("Output", fg, nil).Return("undo_output")

		result := fi.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		fg := newMockStalactitesGame()
		fp := newMockStalactitesPresenter()
		fi := NewStalactitesInteractor(fg, fp)

		err := errors.New("nothing to undo")
		fg.On("Undo").Return(err)
		fp.On("Output", fg, err).Return("error_output")

		result := fi.Undo()
		assert.Equal(t, "error_output", result)
	})
}

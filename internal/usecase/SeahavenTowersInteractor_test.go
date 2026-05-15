package usecase

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockSeahavenTowersGame() *interfaces.MockSeahavenTowersGame {
	return new(interfaces.MockSeahavenTowersGame)
}

func newMockSeahavenTowersPresenter() *presenter.MockSeahavenTowersPresenter {
	return new(presenter.MockSeahavenTowersPresenter)
}

func TestNewSeahavenTowersInteractor(t *testing.T) {
	sg := newMockSeahavenTowersGame()
	sp := newMockSeahavenTowersPresenter()
	si := NewSeahavenTowersInteractor(sg, sp)
	assert.NotNil(t, si)
}

func TestNewSeahavenTowersInteractorPanicsOnNil(t *testing.T) {
	sp := newMockSeahavenTowersPresenter()
	assert.Panics(t, func() { NewSeahavenTowersInteractor(nil, sp) })
	sg := newMockSeahavenTowersGame()
	assert.Panics(t, func() { NewSeahavenTowersInteractor(sg, nil) })
}

func TestSeahavenTowersInteractorReset(t *testing.T) {
	sg := newMockSeahavenTowersGame()
	sp := newMockSeahavenTowersPresenter()
	si := NewSeahavenTowersInteractor(sg, sp)

	sg.On("Reset").Return()
	sp.On("Output", sg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", si.Reset())
	sg.AssertCalled(t, "Reset")
}

func TestSeahavenTowersInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		sg.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		sp.On("Output", sg, nil).Return("move_output")

		assert.Equal(t, "move_output", si.MoveTableauToTableau(0, 2, 3))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		err := errors.New("invalid")
		sg.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		sp.On("Output", sg, err).Return("error_output")

		assert.Equal(t, "error_output", si.MoveTableauToTableau(0, 2, 3))
	})
}

func TestSeahavenTowersInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		sg.On("MoveTableauToFoundation", 2).Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.MoveTableauToFoundation(2))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		err := errors.New("bad")
		sg.On("MoveTableauToFoundation", 2).Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.MoveTableauToFoundation(2))
	})
}

func TestSeahavenTowersInteractorMoveTableauToFreeCell(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		sg.On("MoveTableauToFreeCell", 0, 1).Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.MoveTableauToFreeCell(0, 1))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		err := errors.New("occupied")
		sg.On("MoveTableauToFreeCell", 0, 1).Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.MoveTableauToFreeCell(0, 1))
	})
}

func TestSeahavenTowersInteractorMoveFreeCellToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		sg.On("MoveFreeCellToTableau", 1, 3).Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.MoveFreeCellToTableau(1, 3))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		err := errors.New("empty")
		sg.On("MoveFreeCellToTableau", 1, 3).Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.MoveFreeCellToTableau(1, 3))
	})
}

func TestSeahavenTowersInteractorMoveFreeCellToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		sg.On("MoveFreeCellToFoundation", 0).Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.MoveFreeCellToFoundation(0))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		err := errors.New("nope")
		sg.On("MoveFreeCellToFoundation", 0).Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.MoveFreeCellToFoundation(0))
	})
}

func TestSeahavenTowersInteractorGiveUp(t *testing.T) {
	sg := newMockSeahavenTowersGame()
	sp := newMockSeahavenTowersPresenter()
	si := NewSeahavenTowersInteractor(sg, sp)

	sg.On("GiveUp").Return()
	sp.On("Output", sg, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", si.GiveUp())
	sg.AssertCalled(t, "GiveUp")
}

func TestSeahavenTowersInteractorHint(t *testing.T) {
	sg := newMockSeahavenTowersGame()
	sp := newMockSeahavenTowersPresenter()
	si := NewSeahavenTowersInteractor(sg, sp)

	sp.On("HintOutput", mock.Anything).Return("hint_output")
	assert.Equal(t, "hint_output", si.Hint())
}

func TestSeahavenTowersInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		sg.On("AutoComplete").Return(nil)
		sp.On("Output", sg, nil).Return("ok")
		assert.Equal(t, "ok", si.AutoComplete())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		err := errors.New("not playing")
		sg.On("AutoComplete").Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.AutoComplete())
	})
}

func TestSeahavenTowersInteractorActionLog(t *testing.T) {
	sg := newMockSeahavenTowersGame()
	sp := newMockSeahavenTowersPresenter()
	si := NewSeahavenTowersInteractor(sg, sp)

	sp.On("ActionLogOutput", mock.Anything).Return("log_output")
	assert.Equal(t, "log_output", si.ActionLog())
}

func TestSeahavenTowersInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		sg.On("Undo").Return(nil)
		sp.On("Output", sg, nil).Return("undo_output")
		assert.Equal(t, "undo_output", si.Undo())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		err := errors.New("nothing to undo")
		sg.On("Undo").Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.Undo())
	})
}

func TestSeahavenTowersInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		sg.On("UndoN", 3).Return(nil)
		sp.On("Output", sg, nil).Return("undoN_output")
		assert.Equal(t, "undoN_output", si.UndoN(3))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSeahavenTowersGame()
		sp := newMockSeahavenTowersPresenter()
		si := NewSeahavenTowersInteractor(sg, sp)

		err := errors.New("undo step failed")
		sg.On("UndoN", 5).Return(err)
		sp.On("Output", sg, err).Return("err")
		assert.Equal(t, "err", si.UndoN(5))
	})
}

func TestSeahavenTowersInteractorSnapshotRoundTrip(t *testing.T) {
	// Snapshot must serialise into a payload that RestoreSeahavenTowersInteractor
	// can deserialise into a working interactor. The usecase layer cannot import
	// the adapter presenter implementations (Clean Architecture); use the mock
	// presenter with permissive setups instead.
	original := domain.NewDefaultSeahavenTowers()
	original.Reset()
	sp1 := newMockSeahavenTowersPresenter()
	sp1.On("Output", mock.Anything, mock.Anything).Return("ok").Maybe()
	si := NewSeahavenTowersInteractor(original, sp1)

	data, err := si.Snapshot()
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.True(t, json.Valid(data), "snapshot must be valid JSON")

	sp2 := newMockSeahavenTowersPresenter()
	sp2.On("HintOutput", mock.Anything).Return("hint").Maybe()
	sp2.On("ActionLogOutput", mock.Anything).Return("log").Maybe()
	sp2.On("Output", mock.Anything, mock.Anything).Return("ok").Maybe()

	restored, err := RestoreSeahavenTowersInteractor(data, sp2)
	require.NoError(t, err)
	require.NotNil(t, restored)

	assert.Equal(t, "hint", restored.Hint())
	assert.Equal(t, "log", restored.ActionLog())
}

func TestRestoreSeahavenTowersInteractorRejectsBadJSON(t *testing.T) {
	sp := newMockSeahavenTowersPresenter()
	_, err := RestoreSeahavenTowersInteractor([]byte("not-json"), sp)
	assert.Error(t, err)
}

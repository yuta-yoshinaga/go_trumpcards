package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockKingAlbertGame() *interfaces.MockKingAlbertGame {
	return new(interfaces.MockKingAlbertGame)
}

func newMockKingAlbertPresenter() *presenter.MockKingAlbertPresenter {
	return new(presenter.MockKingAlbertPresenter)
}

func TestNewKingAlbertInteractor(t *testing.T) {
	bg := newMockKingAlbertGame()
	bp := newMockKingAlbertPresenter()
	bi := NewKingAlbertInteractor(bg, bp)
	assert.NotNil(t, bi)
}

func TestNewKingAlbertInteractorPanicsOnNil(t *testing.T) {
	bp := newMockKingAlbertPresenter()
	assert.Panics(t, func() { NewKingAlbertInteractor(nil, bp) })
	bg := newMockKingAlbertGame()
	assert.Panics(t, func() { NewKingAlbertInteractor(bg, nil) })
}

func TestKingAlbertInteractorReset(t *testing.T) {
	bg := newMockKingAlbertGame()
	bp := newMockKingAlbertPresenter()
	bi := NewKingAlbertInteractor(bg, bp)

	bg.On("Reset").Return()
	bp.On("Output", bg, nil).Return("reset_output")

	result := bi.Reset()
	assert.Equal(t, "reset_output", result)
	bg.AssertCalled(t, "Reset")
}

func TestKingAlbertInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		bg.On("MoveTableauToTableau", 0, 3, 5).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToTableau", 0, 3, 5).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "error_output", result)
	})
}

func TestKingAlbertInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		bg.On("MoveTableauToFoundation", 2).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToFoundation", 2).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestKingAlbertInteractorMoveReserveToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		bg.On("MoveReserveToTableau", 1, 4).Return(nil)
		bp.On("Output", bg, nil).Return("rt_output")

		result := bi.MoveReserveToTableau(1, 4)
		assert.Equal(t, "rt_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveReserveToTableau", 1, 4).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveReserveToTableau(1, 4)
		assert.Equal(t, "error_output", result)
	})
}

func TestKingAlbertInteractorMoveReserveToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		bg.On("MoveReserveToFoundation", 0).Return(nil)
		bp.On("Output", bg, nil).Return("rf_output")

		result := bi.MoveReserveToFoundation(0)
		assert.Equal(t, "rf_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveReserveToFoundation", 0).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveReserveToFoundation(0)
		assert.Equal(t, "error_output", result)
	})
}

func TestKingAlbertInteractorGiveUp(t *testing.T) {
	bg := newMockKingAlbertGame()
	bp := newMockKingAlbertPresenter()
	bi := NewKingAlbertInteractor(bg, bp)

	bg.On("GiveUp").Return()
	bp.On("Output", bg, nil).Return("giveup_output")

	result := bi.GiveUp()
	assert.Equal(t, "giveup_output", result)
	bg.AssertCalled(t, "GiveUp")
}

func TestKingAlbertInteractorHint(t *testing.T) {
	bg := newMockKingAlbertGame()
	bp := newMockKingAlbertPresenter()
	bi := NewKingAlbertInteractor(bg, bp)

	bp.On("HintOutput", mock.Anything).Return("hint_output")

	result := bi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestKingAlbertInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		bg.On("AutoComplete").Return(nil)
		bp.On("Output", bg, nil).Return("ac_output")

		result := bi.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		err := errors.New("not playing")
		bg.On("AutoComplete").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestKingAlbertInteractorActionLog(t *testing.T) {
	bg := newMockKingAlbertGame()
	bp := newMockKingAlbertPresenter()
	bi := NewKingAlbertInteractor(bg, bp)

	bp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := bi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestKingAlbertInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		bg.On("Undo").Return(nil)
		bp.On("Output", bg, nil).Return("undo_output")

		result := bi.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		err := errors.New("nothing to undo")
		bg.On("Undo").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.Undo()
		assert.Equal(t, "error_output", result)
	})
}

func TestKingAlbertInteractorSnapshot(t *testing.T) {
	bd := domain.NewDefaultKingAlbert()
	bd.Reset()
	bp := newMockKingAlbertPresenter()
	bi := NewKingAlbertInteractor(bd, bp)

	data, err := bi.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreKingAlbertInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		bd := domain.NewDefaultKingAlbert()
		bd.Reset()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bd, bp)
		data, err := bi.Snapshot()
		require.NoError(t, err)

		restored, err := RestoreKingAlbertInteractor(data, bp)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		bp := newMockKingAlbertPresenter()
		_, err := RestoreKingAlbertInteractor([]byte("not json"), bp)
		assert.Error(t, err)
	})
}

func TestKingAlbertInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		bg.On("UndoN", 3).Return(nil)
		bp.On("Output", bg, nil).Return("undon_output")

		result := bi.UndoN(3)
		assert.Equal(t, "undon_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockKingAlbertGame()
		bp := newMockKingAlbertPresenter()
		bi := NewKingAlbertInteractor(bg, bp)

		err := errors.New("undo failed")
		bg.On("UndoN", 3).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.UndoN(3)
		assert.Equal(t, "error_output", result)
	})
}

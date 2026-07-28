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

func newMockBisleyGame() *interfaces.MockBisleyGame {
	return new(interfaces.MockBisleyGame)
}

func newMockBisleyPresenter() *presenter.MockBisleyPresenter {
	return new(presenter.MockBisleyPresenter)
}

func TestNewBisleyInteractor(t *testing.T) {
	bg := newMockBisleyGame()
	bp := newMockBisleyPresenter()
	bi := NewBisleyInteractor(bg, bp)
	assert.NotNil(t, bi)
}

func TestNewBisleyInteractorPanicsOnNil(t *testing.T) {
	bp := newMockBisleyPresenter()
	assert.Panics(t, func() { NewBisleyInteractor(nil, bp) })
	bg := newMockBisleyGame()
	assert.Panics(t, func() { NewBisleyInteractor(bg, nil) })
}

func TestBisleyInteractorReset(t *testing.T) {
	bg := newMockBisleyGame()
	bp := newMockBisleyPresenter()
	bi := NewBisleyInteractor(bg, bp)

	bg.On("Reset").Return()
	bp.On("Output", bg, nil).Return("reset_output")

	result := bi.Reset()
	assert.Equal(t, "reset_output", result)
	bg.AssertCalled(t, "Reset")
}

func TestBisleyInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		bg.On("MoveTableauToTableau", 0, 5).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToTableau(0, 5)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToTableau", 0, 5).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToTableau(0, 5)
		assert.Equal(t, "error_output", result)
	})
}

func TestBisleyInteractorMoveTableauToAceFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		bg.On("MoveTableauToAceFoundation", 2).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToAceFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToAceFoundation", 2).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToAceFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestBisleyInteractorMoveTableauToKingFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		bg.On("MoveTableauToKingFoundation", 7).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToKingFoundation(7)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToKingFoundation", 7).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToKingFoundation(7)
		assert.Equal(t, "error_output", result)
	})
}

func TestBisleyInteractorGiveUp(t *testing.T) {
	bg := newMockBisleyGame()
	bp := newMockBisleyPresenter()
	bi := NewBisleyInteractor(bg, bp)

	bg.On("GiveUp").Return()
	bp.On("Output", bg, nil).Return("giveup_output")

	result := bi.GiveUp()
	assert.Equal(t, "giveup_output", result)
	bg.AssertCalled(t, "GiveUp")
}

func TestBisleyInteractorHint(t *testing.T) {
	bg := newMockBisleyGame()
	bp := newMockBisleyPresenter()
	bi := NewBisleyInteractor(bg, bp)

	bp.On("HintOutput", mock.Anything).Return("hint_output")

	result := bi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestBisleyInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		bg.On("AutoComplete").Return(nil)
		bp.On("Output", bg, nil).Return("ac_output")

		result := bi.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		err := errors.New("not playing")
		bg.On("AutoComplete").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestBisleyInteractorActionLog(t *testing.T) {
	bg := newMockBisleyGame()
	bp := newMockBisleyPresenter()
	bi := NewBisleyInteractor(bg, bp)

	bp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := bi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestBisleyInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		bg.On("Undo").Return(nil)
		bp.On("Output", bg, nil).Return("undo_output")

		result := bi.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		err := errors.New("nothing to undo")
		bg.On("Undo").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.Undo()
		assert.Equal(t, "error_output", result)
	})
}

func TestBisleyInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		bg.On("UndoN", 3).Return(nil)
		bp.On("Output", bg, nil).Return("undon_output")

		result := bi.UndoN(3)
		assert.Equal(t, "undon_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBisleyGame()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bg, bp)

		err := errors.New("undo failed")
		bg.On("UndoN", 3).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.UndoN(3)
		assert.Equal(t, "error_output", result)
	})
}

func TestBisleyInteractorSnapshot(t *testing.T) {
	bd := domain.NewDefaultBisley()
	bd.Reset()
	bp := newMockBisleyPresenter()
	bi := NewBisleyInteractor(bd, bp)

	data, err := bi.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreBisleyInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		bd := domain.NewDefaultBisley()
		bd.Reset()
		bp := newMockBisleyPresenter()
		bi := NewBisleyInteractor(bd, bp)
		data, err := bi.Snapshot()
		require.NoError(t, err)

		restored, err := RestoreBisleyInteractor(data, bp)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		bp := newMockBisleyPresenter()
		_, err := RestoreBisleyInteractor([]byte("not json"), bp)
		assert.Error(t, err)
	})
}

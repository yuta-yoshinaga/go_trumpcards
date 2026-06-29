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

func newMockFlowerGardenGame() *interfaces.MockFlowerGardenGame {
	return new(interfaces.MockFlowerGardenGame)
}

func newMockFlowerGardenPresenter() *presenter.MockFlowerGardenPresenter {
	return new(presenter.MockFlowerGardenPresenter)
}

func TestNewFlowerGardenInteractor(t *testing.T) {
	bg := newMockFlowerGardenGame()
	bp := newMockFlowerGardenPresenter()
	bi := NewFlowerGardenInteractor(bg, bp)
	assert.NotNil(t, bi)
}

func TestNewFlowerGardenInteractorPanicsOnNil(t *testing.T) {
	bp := newMockFlowerGardenPresenter()
	assert.Panics(t, func() { NewFlowerGardenInteractor(nil, bp) })
	bg := newMockFlowerGardenGame()
	assert.Panics(t, func() { NewFlowerGardenInteractor(bg, nil) })
}

func TestFlowerGardenInteractorReset(t *testing.T) {
	bg := newMockFlowerGardenGame()
	bp := newMockFlowerGardenPresenter()
	bi := NewFlowerGardenInteractor(bg, bp)

	bg.On("Reset").Return()
	bp.On("Output", bg, nil).Return("reset_output")

	result := bi.Reset()
	assert.Equal(t, "reset_output", result)
	bg.AssertCalled(t, "Reset")
}

func TestFlowerGardenInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		bg.On("MoveTableauToTableau", 0, 3, 5).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToTableau", 0, 3, 5).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "error_output", result)
	})
}

func TestFlowerGardenInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		bg.On("MoveTableauToFoundation", 2).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToFoundation", 2).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestFlowerGardenInteractorMoveReserveToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		bg.On("MoveReserveToTableau", 1, 4).Return(nil)
		bp.On("Output", bg, nil).Return("rt_output")

		result := bi.MoveReserveToTableau(1, 4)
		assert.Equal(t, "rt_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveReserveToTableau", 1, 4).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveReserveToTableau(1, 4)
		assert.Equal(t, "error_output", result)
	})
}

func TestFlowerGardenInteractorMoveReserveToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		bg.On("MoveReserveToFoundation", 0).Return(nil)
		bp.On("Output", bg, nil).Return("rf_output")

		result := bi.MoveReserveToFoundation(0)
		assert.Equal(t, "rf_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveReserveToFoundation", 0).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveReserveToFoundation(0)
		assert.Equal(t, "error_output", result)
	})
}

func TestFlowerGardenInteractorGiveUp(t *testing.T) {
	bg := newMockFlowerGardenGame()
	bp := newMockFlowerGardenPresenter()
	bi := NewFlowerGardenInteractor(bg, bp)

	bg.On("GiveUp").Return()
	bp.On("Output", bg, nil).Return("giveup_output")

	result := bi.GiveUp()
	assert.Equal(t, "giveup_output", result)
	bg.AssertCalled(t, "GiveUp")
}

func TestFlowerGardenInteractorHint(t *testing.T) {
	bg := newMockFlowerGardenGame()
	bp := newMockFlowerGardenPresenter()
	bi := NewFlowerGardenInteractor(bg, bp)

	bp.On("HintOutput", mock.Anything).Return("hint_output")

	result := bi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestFlowerGardenInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		bg.On("AutoComplete").Return(nil)
		bp.On("Output", bg, nil).Return("ac_output")

		result := bi.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		err := errors.New("not playing")
		bg.On("AutoComplete").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestFlowerGardenInteractorActionLog(t *testing.T) {
	bg := newMockFlowerGardenGame()
	bp := newMockFlowerGardenPresenter()
	bi := NewFlowerGardenInteractor(bg, bp)

	bp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := bi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestFlowerGardenInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		bg.On("Undo").Return(nil)
		bp.On("Output", bg, nil).Return("undo_output")

		result := bi.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		err := errors.New("nothing to undo")
		bg.On("Undo").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.Undo()
		assert.Equal(t, "error_output", result)
	})
}

func TestFlowerGardenInteractorSnapshot(t *testing.T) {
	bd := domain.NewDefaultFlowerGarden()
	bd.Reset()
	bp := newMockFlowerGardenPresenter()
	bi := NewFlowerGardenInteractor(bd, bp)

	data, err := bi.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreFlowerGardenInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		bd := domain.NewDefaultFlowerGarden()
		bd.Reset()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bd, bp)
		data, err := bi.Snapshot()
		require.NoError(t, err)

		restored, err := RestoreFlowerGardenInteractor(data, bp)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		bp := newMockFlowerGardenPresenter()
		_, err := RestoreFlowerGardenInteractor([]byte("not json"), bp)
		assert.Error(t, err)
	})
}

func TestFlowerGardenInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		bg.On("UndoN", 3).Return(nil)
		bp.On("Output", bg, nil).Return("undon_output")

		result := bi.UndoN(3)
		assert.Equal(t, "undon_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFlowerGardenGame()
		bp := newMockFlowerGardenPresenter()
		bi := NewFlowerGardenInteractor(bg, bp)

		err := errors.New("undo failed")
		bg.On("UndoN", 3).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.UndoN(3)
		assert.Equal(t, "error_output", result)
	})
}

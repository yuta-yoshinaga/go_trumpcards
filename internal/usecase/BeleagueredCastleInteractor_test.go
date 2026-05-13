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

func newMockBeleagueredCastleGame() *interfaces.MockBeleagueredCastleGame {
	return new(interfaces.MockBeleagueredCastleGame)
}

func newMockBeleagueredCastlePresenter() *presenter.MockBeleagueredCastlePresenter {
	return new(presenter.MockBeleagueredCastlePresenter)
}

func TestNewBeleagueredCastleInteractor(t *testing.T) {
	bg := newMockBeleagueredCastleGame()
	bp := newMockBeleagueredCastlePresenter()
	bi := NewBeleagueredCastleInteractor(bg, bp)
	assert.NotNil(t, bi)
}

func TestNewBeleagueredCastleInteractorPanicsOnNil(t *testing.T) {
	bp := newMockBeleagueredCastlePresenter()
	assert.Panics(t, func() { NewBeleagueredCastleInteractor(nil, bp) })
	bg := newMockBeleagueredCastleGame()
	assert.Panics(t, func() { NewBeleagueredCastleInteractor(bg, nil) })
}

func TestBeleagueredCastleInteractorReset(t *testing.T) {
	bg := newMockBeleagueredCastleGame()
	bp := newMockBeleagueredCastlePresenter()
	bi := NewBeleagueredCastleInteractor(bg, bp)

	bg.On("Reset").Return()
	bp.On("Output", bg, nil).Return("reset_output")

	result := bi.Reset()
	assert.Equal(t, "reset_output", result)
	bg.AssertCalled(t, "Reset")
}

func TestBeleagueredCastleInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		bg.On("MoveTableauToTableau", 0, 3, 5).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToTableau", 0, 3, 5).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "error_output", result)
	})
}

func TestBeleagueredCastleInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		bg.On("MoveTableauToFoundation", 2).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToFoundation", 2).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestBeleagueredCastleInteractorGiveUp(t *testing.T) {
	bg := newMockBeleagueredCastleGame()
	bp := newMockBeleagueredCastlePresenter()
	bi := NewBeleagueredCastleInteractor(bg, bp)

	bg.On("GiveUp").Return()
	bp.On("Output", bg, nil).Return("giveup_output")

	result := bi.GiveUp()
	assert.Equal(t, "giveup_output", result)
	bg.AssertCalled(t, "GiveUp")
}

func TestBeleagueredCastleInteractorHint(t *testing.T) {
	bg := newMockBeleagueredCastleGame()
	bp := newMockBeleagueredCastlePresenter()
	bi := NewBeleagueredCastleInteractor(bg, bp)

	bp.On("HintOutput", mock.Anything).Return("hint_output")

	result := bi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestBeleagueredCastleInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		bg.On("AutoComplete").Return(nil)
		bp.On("Output", bg, nil).Return("ac_output")

		result := bi.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		err := errors.New("not playing")
		bg.On("AutoComplete").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestBeleagueredCastleInteractorActionLog(t *testing.T) {
	bg := newMockBeleagueredCastleGame()
	bp := newMockBeleagueredCastlePresenter()
	bi := NewBeleagueredCastleInteractor(bg, bp)

	bp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := bi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestBeleagueredCastleInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		bg.On("Undo").Return(nil)
		bp.On("Output", bg, nil).Return("undo_output")

		result := bi.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		err := errors.New("nothing to undo")
		bg.On("Undo").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.Undo()
		assert.Equal(t, "error_output", result)
	})
}

func TestBeleagueredCastleInteractorSnapshot(t *testing.T) {
	bd := domain.NewDefaultBeleagueredCastle()
	bd.Reset()
	bp := newMockBeleagueredCastlePresenter()
	bi := NewBeleagueredCastleInteractor(bd, bp)

	data, err := bi.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreBeleagueredCastleInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		bd := domain.NewDefaultBeleagueredCastle()
		bd.Reset()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bd, bp)
		data, err := bi.Snapshot()
		require.NoError(t, err)

		restored, err := RestoreBeleagueredCastleInteractor(data, bp)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		bp := newMockBeleagueredCastlePresenter()
		_, err := RestoreBeleagueredCastleInteractor([]byte("not json"), bp)
		assert.Error(t, err)
	})
}

func TestBeleagueredCastleInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		bg.On("UndoN", 3).Return(nil)
		bp.On("Output", bg, nil).Return("undon_output")

		result := bi.UndoN(3)
		assert.Equal(t, "undon_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockBeleagueredCastleGame()
		bp := newMockBeleagueredCastlePresenter()
		bi := NewBeleagueredCastleInteractor(bg, bp)

		err := errors.New("undo failed")
		bg.On("UndoN", 3).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.UndoN(3)
		assert.Equal(t, "error_output", result)
	})
}

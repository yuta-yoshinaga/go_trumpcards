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

func newMockFortressGame() *interfaces.MockFortressGame {
	return new(interfaces.MockFortressGame)
}

func newMockFortressPresenter() *presenter.MockFortressPresenter {
	return new(presenter.MockFortressPresenter)
}

func TestNewFortressInteractor(t *testing.T) {
	bg := newMockFortressGame()
	bp := newMockFortressPresenter()
	bi := NewFortressInteractor(bg, bp)
	assert.NotNil(t, bi)
}

func TestNewFortressInteractorPanicsOnNil(t *testing.T) {
	bp := newMockFortressPresenter()
	assert.Panics(t, func() { NewFortressInteractor(nil, bp) })
	bg := newMockFortressGame()
	assert.Panics(t, func() { NewFortressInteractor(bg, nil) })
}

func TestFortressInteractorReset(t *testing.T) {
	bg := newMockFortressGame()
	bp := newMockFortressPresenter()
	bi := NewFortressInteractor(bg, bp)

	bg.On("Reset").Return()
	bp.On("Output", bg, nil).Return("reset_output")

	result := bi.Reset()
	assert.Equal(t, "reset_output", result)
	bg.AssertCalled(t, "Reset")
}

func TestFortressInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		bg.On("MoveTableauToTableau", 0, 3, 5).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToTableau", 0, 3, 5).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToTableau(0, 3, 5)
		assert.Equal(t, "error_output", result)
	})
}

func TestFortressInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		bg.On("MoveTableauToFoundation", 2).Return(nil)
		bp.On("Output", bg, nil).Return("move_output")

		result := bi.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		err := errors.New("invalid")
		bg.On("MoveTableauToFoundation", 2).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestFortressInteractorGiveUp(t *testing.T) {
	bg := newMockFortressGame()
	bp := newMockFortressPresenter()
	bi := NewFortressInteractor(bg, bp)

	bg.On("GiveUp").Return()
	bp.On("Output", bg, nil).Return("giveup_output")

	result := bi.GiveUp()
	assert.Equal(t, "giveup_output", result)
	bg.AssertCalled(t, "GiveUp")
}

func TestFortressInteractorHint(t *testing.T) {
	bg := newMockFortressGame()
	bp := newMockFortressPresenter()
	bi := NewFortressInteractor(bg, bp)

	bp.On("HintOutput", mock.Anything).Return("hint_output")

	result := bi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestFortressInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		bg.On("AutoComplete").Return(nil)
		bp.On("Output", bg, nil).Return("ac_output")

		result := bi.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		err := errors.New("not playing")
		bg.On("AutoComplete").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestFortressInteractorActionLog(t *testing.T) {
	bg := newMockFortressGame()
	bp := newMockFortressPresenter()
	bi := NewFortressInteractor(bg, bp)

	bp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := bi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestFortressInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		bg.On("Undo").Return(nil)
		bp.On("Output", bg, nil).Return("undo_output")

		result := bi.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		err := errors.New("nothing to undo")
		bg.On("Undo").Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.Undo()
		assert.Equal(t, "error_output", result)
	})
}

func TestFortressInteractorSnapshot(t *testing.T) {
	bd := domain.NewDefaultFortress()
	bd.Reset()
	bp := newMockFortressPresenter()
	bi := NewFortressInteractor(bd, bp)

	data, err := bi.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreFortressInteractor(t *testing.T) {
	t.Run("round-trip preserves state", func(t *testing.T) {
		bd := domain.NewDefaultFortress()
		bd.Reset()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bd, bp)
		data, err := bi.Snapshot()
		require.NoError(t, err)

		restored, err := RestoreFortressInteractor(data, bp)
		require.NoError(t, err)
		assert.NotNil(t, restored)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		bp := newMockFortressPresenter()
		_, err := RestoreFortressInteractor([]byte("not json"), bp)
		assert.Error(t, err)
	})
}

func TestFortressInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		bg.On("UndoN", 3).Return(nil)
		bp.On("Output", bg, nil).Return("undon_output")

		result := bi.UndoN(3)
		assert.Equal(t, "undon_output", result)
	})

	t.Run("error", func(t *testing.T) {
		bg := newMockFortressGame()
		bp := newMockFortressPresenter()
		bi := NewFortressInteractor(bg, bp)

		err := errors.New("undo failed")
		bg.On("UndoN", 3).Return(err)
		bp.On("Output", bg, err).Return("error_output")

		result := bi.UndoN(3)
		assert.Equal(t, "error_output", result)
	})
}

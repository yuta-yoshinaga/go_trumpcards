package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockEasthavenGame() *interfaces.MockEasthavenGame {
	return new(interfaces.MockEasthavenGame)
}

func newMockEasthavenPresenter() *presenter.MockEasthavenPresenter {
	return new(presenter.MockEasthavenPresenter)
}

func TestNewEasthavenInteractor(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)
	assert.NotNil(t, ei)
}

func TestNewEasthavenInteractorPanicsOnNil(t *testing.T) {
	ep := newMockEasthavenPresenter()
	assert.Panics(t, func() { NewEasthavenInteractor(nil, ep) })
	eg := newMockEasthavenGame()
	assert.Panics(t, func() { NewEasthavenInteractor(eg, nil) })
}

func TestEasthavenInteractorReset(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)

	eg.On("Reset").Return()
	ep.On("Output", eg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", ei.Reset())
	eg.AssertCalled(t, "Reset")
}

func TestEasthavenInteractorDeal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		eg := newMockEasthavenGame()
		ep := newMockEasthavenPresenter()
		ei := NewEasthavenInteractor(eg, ep)

		eg.On("Deal").Return(nil)
		ep.On("Output", eg, nil).Return("deal_output")

		assert.Equal(t, "deal_output", ei.Deal())
	})

	t.Run("error", func(t *testing.T) {
		eg := newMockEasthavenGame()
		ep := newMockEasthavenPresenter()
		ei := NewEasthavenInteractor(eg, ep)

		err := errors.New("cannot deal")
		eg.On("Deal").Return(err)
		ep.On("Output", eg, err).Return("deal_err")

		assert.Equal(t, "deal_err", ei.Deal())
	})
}

func TestEasthavenInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		eg := newMockEasthavenGame()
		ep := newMockEasthavenPresenter()
		ei := NewEasthavenInteractor(eg, ep)

		eg.On("MoveTableauToTableau", 0, 1, 2).Return(nil)
		ep.On("Output", eg, nil).Return("move_output")

		assert.Equal(t, "move_output", ei.MoveTableauToTableau(0, 1, 2))
	})

	t.Run("error", func(t *testing.T) {
		eg := newMockEasthavenGame()
		ep := newMockEasthavenPresenter()
		ei := NewEasthavenInteractor(eg, ep)

		err := errors.New("invalid move")
		eg.On("MoveTableauToTableau", 0, 1, 2).Return(err)
		ep.On("Output", eg, err).Return("error_output")

		assert.Equal(t, "error_output", ei.MoveTableauToTableau(0, 1, 2))
	})
}

func TestEasthavenInteractorMoveTableauToFoundation(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)

	eg.On("MoveTableauToFoundation", 3).Return(nil)
	ep.On("Output", eg, nil).Return("move_f_output")

	assert.Equal(t, "move_f_output", ei.MoveTableauToFoundation(3))
}

func TestEasthavenInteractorGiveUp(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)

	eg.On("GiveUp").Return()
	ep.On("Output", eg, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", ei.GiveUp())
}

func TestEasthavenInteractorHint(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)

	ep.On("HintOutput", eg).Return("hint_output")

	assert.Equal(t, "hint_output", ei.Hint())
}

func TestEasthavenInteractorAutoComplete(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)

	eg.On("AutoComplete").Return(nil)
	ep.On("Output", eg, nil).Return("ac_output")

	assert.Equal(t, "ac_output", ei.AutoComplete())
}

func TestEasthavenInteractorActionLog(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)

	ep.On("ActionLogOutput", eg).Return("log_output")

	assert.Equal(t, "log_output", ei.ActionLog())
}

func TestEasthavenInteractorUndo(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)

	eg.On("Undo").Return(nil)
	ep.On("Output", eg, nil).Return("undo_output")

	assert.Equal(t, "undo_output", ei.Undo())
}

func TestEasthavenInteractorUndoN(t *testing.T) {
	eg := newMockEasthavenGame()
	ep := newMockEasthavenPresenter()
	ei := NewEasthavenInteractor(eg, ep)

	eg.On("UndoN", 3).Return(nil)
	ep.On("Output", eg, nil).Return("undo_n_output")

	assert.Equal(t, "undo_n_output", ei.UndoN(3))
}

func TestRestoreEasthavenInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"tb":[[],[],[],[],[],[],[]],"st":[],"fd":[[],[],[],[]],"ps":0,"mc":0,"al":[],"sl":false}`)
		ep := newMockEasthavenPresenter()
		ei, err := RestoreEasthavenInteractor(data, ep)
		assert.NoError(t, err)
		assert.NotNil(t, ei)
	})

	t.Run("invalid data", func(t *testing.T) {
		ep := newMockEasthavenPresenter()
		_, err := RestoreEasthavenInteractor([]byte("invalid"), ep)
		assert.Error(t, err)
	})
}

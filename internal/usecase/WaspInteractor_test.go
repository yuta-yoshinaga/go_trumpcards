package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockWaspGame() *interfaces.MockWaspGame {
	return new(interfaces.MockWaspGame)
}

func newMockWaspPresenter() *presenter.MockWaspPresenter {
	return new(presenter.MockWaspPresenter)
}

func TestNewWaspInteractor(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)
	assert.NotNil(t, si)
}

func TestNewWaspInteractorPanicsOnNil(t *testing.T) {
	sp := newMockWaspPresenter()
	assert.Panics(t, func() { NewWaspInteractor(nil, sp) })
	sg := newMockWaspGame()
	assert.Panics(t, func() { NewWaspInteractor(sg, nil) })
}

func TestWaspInteractorReset(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)

	sg.On("Reset").Return()
	sp.On("Output", sg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", si.Reset())
	sg.AssertCalled(t, "Reset")
}

func TestWaspInteractorDeal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockWaspGame()
		sp := newMockWaspPresenter()
		si := NewWaspInteractor(sg, sp)

		sg.On("Deal").Return(nil)
		sp.On("Output", sg, nil).Return("deal_output")

		assert.Equal(t, "deal_output", si.Deal())
	})

	t.Run("error", func(t *testing.T) {
		sg := newMockWaspGame()
		sp := newMockWaspPresenter()
		si := NewWaspInteractor(sg, sp)

		err := errors.New("no stock")
		sg.On("Deal").Return(err)
		sp.On("Output", sg, err).Return("deal_err")

		assert.Equal(t, "deal_err", si.Deal())
	})
}

func TestWaspInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockWaspGame()
		sp := newMockWaspPresenter()
		si := NewWaspInteractor(sg, sp)

		sg.On("MoveTableauToTableau", 0, 1, 2).Return(nil)
		sp.On("Output", sg, nil).Return("move_output")

		assert.Equal(t, "move_output", si.MoveTableauToTableau(0, 1, 2))
	})

	t.Run("error", func(t *testing.T) {
		sg := newMockWaspGame()
		sp := newMockWaspPresenter()
		si := NewWaspInteractor(sg, sp)

		err := errors.New("invalid move")
		sg.On("MoveTableauToTableau", 0, 1, 2).Return(err)
		sp.On("Output", sg, err).Return("error_output")

		assert.Equal(t, "error_output", si.MoveTableauToTableau(0, 1, 2))
	})
}

func TestWaspInteractorGiveUp(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)

	sg.On("GiveUp").Return()
	sp.On("Output", sg, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", si.GiveUp())
}

func TestWaspInteractorHint(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)

	sp.On("HintOutput", sg).Return("hint_output")

	assert.Equal(t, "hint_output", si.Hint())
}

func TestWaspInteractorLegalMoves(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)

	sp.On("LegalMovesOutput", sg, 2).Return("legal_output")

	assert.Equal(t, "legal_output", si.LegalMoves(2))
}

func TestWaspInteractorAutoComplete(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)

	sg.On("AutoComplete").Return(nil)
	sp.On("Output", sg, nil).Return("ac_output")

	assert.Equal(t, "ac_output", si.AutoComplete())
}

func TestWaspInteractorActionLog(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)

	sp.On("ActionLogOutput", sg).Return("log_output")

	assert.Equal(t, "log_output", si.ActionLog())
}

func TestWaspInteractorUndo(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)

	sg.On("Undo").Return(nil)
	sp.On("Output", sg, nil).Return("undo_output")

	assert.Equal(t, "undo_output", si.Undo())
}

func TestWaspInteractorUndoN(t *testing.T) {
	sg := newMockWaspGame()
	sp := newMockWaspPresenter()
	si := NewWaspInteractor(sg, sp)

	sg.On("UndoN", 3).Return(nil)
	sp.On("Output", sg, nil).Return("undo_n_output")

	assert.Equal(t, "undo_n_output", si.UndoN(3))
}

func TestRestoreWaspInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"tb":[[],[],[],[],[],[],[]],"st":[],"cs":0,"ps":0,"mc":0,"al":[],"sl":false}`)
		sp := newMockWaspPresenter()
		si, err := RestoreWaspInteractor(data, sp)
		assert.NoError(t, err)
		assert.NotNil(t, si)
	})

	t.Run("invalid data", func(t *testing.T) {
		sp := newMockWaspPresenter()
		_, err := RestoreWaspInteractor([]byte("invalid"), sp)
		assert.Error(t, err)
	})
}

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockRussianSolitaireGame() *interfaces.MockRussianSolitaireGame {
	return new(interfaces.MockRussianSolitaireGame)
}

func newMockRussianSolitairePresenter() *presenter.MockRussianSolitairePresenter {
	return new(presenter.MockRussianSolitairePresenter)
}

func TestNewRussianSolitaireInteractor(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)
	assert.NotNil(t, ri)
}

func TestNewRussianSolitaireInteractorPanicsOnNil(t *testing.T) {
	rp := newMockRussianSolitairePresenter()
	assert.Panics(t, func() { NewRussianSolitaireInteractor(nil, rp) })
	rg := newMockRussianSolitaireGame()
	assert.Panics(t, func() { NewRussianSolitaireInteractor(rg, nil) })
}

func TestRussianSolitaireInteractorReset(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)

	rg.On("Reset").Return()
	rp.On("Output", rg, nil).Return("reset_output")

	result := ri.Reset()
	assert.Equal(t, "reset_output", result)
	rg.AssertCalled(t, "Reset")
}

func TestRussianSolitaireInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rg := newMockRussianSolitaireGame()
		rp := newMockRussianSolitairePresenter()
		ri := NewRussianSolitaireInteractor(rg, rp)

		rg.On("MoveTableauToTableau", 0, 1, 2).Return(nil)
		rp.On("Output", rg, nil).Return("move_output")

		result := ri.MoveTableauToTableau(0, 1, 2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		rg := newMockRussianSolitaireGame()
		rp := newMockRussianSolitairePresenter()
		ri := NewRussianSolitaireInteractor(rg, rp)

		err := errors.New("invalid move")
		rg.On("MoveTableauToTableau", 0, 1, 2).Return(err)
		rp.On("Output", rg, err).Return("error_output")

		result := ri.MoveTableauToTableau(0, 1, 2)
		assert.Equal(t, "error_output", result)
	})
}

func TestRussianSolitaireInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rg := newMockRussianSolitaireGame()
		rp := newMockRussianSolitairePresenter()
		ri := NewRussianSolitaireInteractor(rg, rp)

		rg.On("MoveTableauToFoundation", 3).Return(nil)
		rp.On("Output", rg, nil).Return("move_f_output")

		result := ri.MoveTableauToFoundation(3)
		assert.Equal(t, "move_f_output", result)
	})
}

func TestRussianSolitaireInteractorGiveUp(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)

	rg.On("GiveUp").Return()
	rp.On("Output", rg, nil).Return("giveup_output")

	result := ri.GiveUp()
	assert.Equal(t, "giveup_output", result)
}

func TestRussianSolitaireInteractorHint(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)

	rp.On("HintOutput", rg).Return("hint_output")

	result := ri.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestRussianSolitaireInteractorAutoComplete(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)

	rg.On("AutoComplete").Return(nil)
	rp.On("Output", rg, nil).Return("ac_output")

	result := ri.AutoComplete()
	assert.Equal(t, "ac_output", result)
}

func TestRussianSolitaireInteractorActionLog(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)

	rp.On("ActionLogOutput", rg).Return("log_output")

	result := ri.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestRussianSolitaireInteractorUndo(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)

	rg.On("Undo").Return(nil)
	rp.On("Output", rg, nil).Return("undo_output")

	result := ri.Undo()
	assert.Equal(t, "undo_output", result)
}

func TestRussianSolitaireInteractorUndoN(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)

	rg.On("UndoN", 3).Return(nil)
	rp.On("Output", rg, nil).Return("undo_n_output")

	result := ri.UndoN(3)
	assert.Equal(t, "undo_n_output", result)
}

func TestRussianSolitaireInteractorSnapshot(t *testing.T) {
	rg := newMockRussianSolitaireGame()
	rp := newMockRussianSolitairePresenter()
	ri := NewRussianSolitaireInteractor(rg, rp)
	assert.NotNil(t, ri)
}

func TestRestoreRussianSolitaireInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"tb":[[],[],[],[],[],[],[]],"fd":[[],[],[],[]],"ps":0,"mc":0,"al":[],"sl":false}`)
		rp := newMockRussianSolitairePresenter()
		ri, err := RestoreRussianSolitaireInteractor(data, rp)
		assert.NoError(t, err)
		assert.NotNil(t, ri)
	})

	t.Run("invalid data", func(t *testing.T) {
		rp := newMockRussianSolitairePresenter()
		_, err := RestoreRussianSolitaireInteractor([]byte("invalid"), rp)
		assert.Error(t, err)
	})
}

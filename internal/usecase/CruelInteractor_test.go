package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockCruelGame() *interfaces.MockCruelGame {
	return new(interfaces.MockCruelGame)
}

func newMockCruelPresenter() *presenter.MockCruelPresenter {
	return new(presenter.MockCruelPresenter)
}

func TestNewCruelInteractor(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)
	assert.NotNil(t, ci)
}

func TestNewCruelInteractorPanicsOnNil(t *testing.T) {
	cp := newMockCruelPresenter()
	assert.Panics(t, func() { NewCruelInteractor(nil, cp) })
	cg := newMockCruelGame()
	assert.Panics(t, func() { NewCruelInteractor(cg, nil) })
}

func TestCruelInteractorReset(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)

	cg.On("Reset").Return()
	cp.On("Output", cg, nil).Return("reset_output")

	result := ci.Reset()
	assert.Equal(t, "reset_output", result)
	cg.AssertCalled(t, "Reset")
}

func TestCruelInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockCruelGame()
		cp := newMockCruelPresenter()
		ci := NewCruelInteractor(cg, cp)

		cg.On("MoveTableauToTableau", 0, 2).Return(nil)
		cp.On("Output", cg, nil).Return("move_output")

		result := ci.MoveTableauToTableau(0, 2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		cg := newMockCruelGame()
		cp := newMockCruelPresenter()
		ci := NewCruelInteractor(cg, cp)

		err := errors.New("invalid move")
		cg.On("MoveTableauToTableau", 0, 2).Return(err)
		cp.On("Output", cg, err).Return("error_output")

		result := ci.MoveTableauToTableau(0, 2)
		assert.Equal(t, "error_output", result)
	})
}

func TestCruelInteractorMoveTableauToFoundation(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)

	cg.On("MoveTableauToFoundation", 3).Return(nil)
	cp.On("Output", cg, nil).Return("move_f_output")

	result := ci.MoveTableauToFoundation(3)
	assert.Equal(t, "move_f_output", result)
}

func TestCruelInteractorShift(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cg := newMockCruelGame()
		cp := newMockCruelPresenter()
		ci := NewCruelInteractor(cg, cp)

		cg.On("Shift").Return(nil)
		cp.On("Output", cg, nil).Return("shift_output")

		result := ci.Shift()
		assert.Equal(t, "shift_output", result)
	})

	t.Run("error", func(t *testing.T) {
		cg := newMockCruelGame()
		cp := newMockCruelPresenter()
		ci := NewCruelInteractor(cg, cp)

		err := errors.New("nothing to shift")
		cg.On("Shift").Return(err)
		cp.On("Output", cg, err).Return("shift_error_output")

		result := ci.Shift()
		assert.Equal(t, "shift_error_output", result)
	})
}

func TestCruelInteractorGiveUp(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)

	cg.On("GiveUp").Return()
	cp.On("Output", cg, nil).Return("giveup_output")

	result := ci.GiveUp()
	assert.Equal(t, "giveup_output", result)
}

func TestCruelInteractorHint(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)

	cp.On("HintOutput", cg).Return("hint_output")

	result := ci.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestCruelInteractorAutoComplete(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)

	cg.On("AutoComplete").Return(nil)
	cp.On("Output", cg, nil).Return("ac_output")

	result := ci.AutoComplete()
	assert.Equal(t, "ac_output", result)
}

func TestCruelInteractorActionLog(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)

	cp.On("ActionLogOutput", cg).Return("log_output")

	result := ci.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestCruelInteractorUndo(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)

	cg.On("Undo").Return(nil)
	cp.On("Output", cg, nil).Return("undo_output")

	result := ci.Undo()
	assert.Equal(t, "undo_output", result)
}

func TestCruelInteractorUndoN(t *testing.T) {
	cg := newMockCruelGame()
	cp := newMockCruelPresenter()
	ci := NewCruelInteractor(cg, cp)

	cg.On("UndoN", 3).Return(nil)
	cp.On("Output", cg, nil).Return("undo_n_output")

	result := ci.UndoN(3)
	assert.Equal(t, "undo_n_output", result)
}

func TestRestoreCruelInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"tb":[[],[],[],[],[],[],[],[],[],[],[],[]],"fd":[[],[],[],[]],"ps":0,"mc":0,"al":[],"sl":false}`)
		cp := newMockCruelPresenter()
		ci, err := RestoreCruelInteractor(data, cp)
		assert.NoError(t, err)
		assert.NotNil(t, ci)
	})

	t.Run("invalid data", func(t *testing.T) {
		cp := newMockCruelPresenter()
		_, err := RestoreCruelInteractor([]byte("invalid"), cp)
		assert.Error(t, err)
	})
}

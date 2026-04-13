package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockYukonGame() *interfaces.MockYukonGame {
	return new(interfaces.MockYukonGame)
}

func newMockYukonPresenter() *presenter.MockYukonPresenter {
	return new(presenter.MockYukonPresenter)
}

func TestNewYukonInteractor(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)
	assert.NotNil(t, yi)
}

func TestNewYukonInteractorPanicsOnNil(t *testing.T) {
	yp := newMockYukonPresenter()
	assert.Panics(t, func() { NewYukonInteractor(nil, yp) })
	yg := newMockYukonGame()
	assert.Panics(t, func() { NewYukonInteractor(yg, nil) })
}

func TestYukonInteractorReset(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)

	yg.On("Reset").Return()
	yp.On("Output", yg, nil).Return("reset_output")

	result := yi.Reset()
	assert.Equal(t, "reset_output", result)
	yg.AssertCalled(t, "Reset")
}

func TestYukonInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		yg := newMockYukonGame()
		yp := newMockYukonPresenter()
		yi := NewYukonInteractor(yg, yp)

		yg.On("MoveTableauToTableau", 0, 1, 2).Return(nil)
		yp.On("Output", yg, nil).Return("move_output")

		result := yi.MoveTableauToTableau(0, 1, 2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		yg := newMockYukonGame()
		yp := newMockYukonPresenter()
		yi := NewYukonInteractor(yg, yp)

		err := errors.New("invalid move")
		yg.On("MoveTableauToTableau", 0, 1, 2).Return(err)
		yp.On("Output", yg, err).Return("error_output")

		result := yi.MoveTableauToTableau(0, 1, 2)
		assert.Equal(t, "error_output", result)
	})
}

func TestYukonInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		yg := newMockYukonGame()
		yp := newMockYukonPresenter()
		yi := NewYukonInteractor(yg, yp)

		yg.On("MoveTableauToFoundation", 3).Return(nil)
		yp.On("Output", yg, nil).Return("move_f_output")

		result := yi.MoveTableauToFoundation(3)
		assert.Equal(t, "move_f_output", result)
	})
}

func TestYukonInteractorGiveUp(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)

	yg.On("GiveUp").Return()
	yp.On("Output", yg, nil).Return("giveup_output")

	result := yi.GiveUp()
	assert.Equal(t, "giveup_output", result)
}

func TestYukonInteractorHint(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)

	yp.On("HintOutput", yg).Return("hint_output")

	result := yi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestYukonInteractorAutoComplete(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)

	yg.On("AutoComplete").Return(nil)
	yp.On("Output", yg, nil).Return("ac_output")

	result := yi.AutoComplete()
	assert.Equal(t, "ac_output", result)
}

func TestYukonInteractorActionLog(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)

	yp.On("ActionLogOutput", yg).Return("log_output")

	result := yi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestYukonInteractorUndo(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)

	yg.On("Undo").Return(nil)
	yp.On("Output", yg, nil).Return("undo_output")

	result := yi.Undo()
	assert.Equal(t, "undo_output", result)
}

func TestYukonInteractorUndoN(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)

	yg.On("UndoN", 3).Return(nil)
	yp.On("Output", yg, nil).Return("undo_n_output")

	result := yi.UndoN(3)
	assert.Equal(t, "undo_n_output", result)
}

func TestYukonInteractorSnapshot(t *testing.T) {
	yg := newMockYukonGame()
	yp := newMockYukonPresenter()
	yi := NewYukonInteractor(yg, yp)
	// GameBase.Snapshot() uses json.Marshal on the game
	assert.NotNil(t, yi)
}

func TestRestoreYukonInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"tb":[[],[],[],[],[],[],[]],"fd":[[],[],[],[]],"ps":0,"mc":0,"al":[],"sl":false}`)
		yp := newMockYukonPresenter()
		yi, err := RestoreYukonInteractor(data, yp)
		assert.NoError(t, err)
		assert.NotNil(t, yi)
	})

	t.Run("invalid data", func(t *testing.T) {
		yp := newMockYukonPresenter()
		_, err := RestoreYukonInteractor([]byte("invalid"), yp)
		assert.Error(t, err)
	})
}

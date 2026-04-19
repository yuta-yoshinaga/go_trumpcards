package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockAccordionGame() *interfaces.MockAccordionGame {
	return new(interfaces.MockAccordionGame)
}

func newMockAccordionPresenter() *presenter.MockAccordionPresenter {
	return new(presenter.MockAccordionPresenter)
}

func TestNewAccordionInteractor(t *testing.T) {
	ag := newMockAccordionGame()
	ap := newMockAccordionPresenter()
	ai := NewAccordionInteractor(ag, ap)
	assert.NotNil(t, ai)
}

func TestNewAccordionInteractorPanicsOnNil(t *testing.T) {
	ap := newMockAccordionPresenter()
	assert.Panics(t, func() { NewAccordionInteractor(nil, ap) })
	ag := newMockAccordionGame()
	assert.Panics(t, func() { NewAccordionInteractor(ag, nil) })
}

func TestAccordionInteractorReset(t *testing.T) {
	ag := newMockAccordionGame()
	ap := newMockAccordionPresenter()
	ai := NewAccordionInteractor(ag, ap)

	ag.On("Reset").Return()
	ap.On("Output", ag, nil).Return("reset_output")

	assert.Equal(t, "reset_output", ai.Reset())
	ag.AssertCalled(t, "Reset")
}

func TestAccordionInteractorMove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ag := newMockAccordionGame()
		ap := newMockAccordionPresenter()
		ai := NewAccordionInteractor(ag, ap)

		ag.On("Move", 1, 0).Return(nil)
		ap.On("Output", ag, nil).Return("move_output")

		assert.Equal(t, "move_output", ai.Move(1, 0))
	})

	t.Run("error", func(t *testing.T) {
		ag := newMockAccordionGame()
		ap := newMockAccordionPresenter()
		ai := NewAccordionInteractor(ag, ap)

		err := errors.New("invalid move")
		ag.On("Move", 1, 0).Return(err)
		ap.On("Output", ag, err).Return("error_output")

		assert.Equal(t, "error_output", ai.Move(1, 0))
	})
}

func TestAccordionInteractorGiveUp(t *testing.T) {
	ag := newMockAccordionGame()
	ap := newMockAccordionPresenter()
	ai := NewAccordionInteractor(ag, ap)

	ag.On("GiveUp").Return()
	ap.On("Output", ag, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", ai.GiveUp())
}

func TestAccordionInteractorHint(t *testing.T) {
	ag := newMockAccordionGame()
	ap := newMockAccordionPresenter()
	ai := NewAccordionInteractor(ag, ap)

	ap.On("HintOutput", ag).Return("hint_output")

	assert.Equal(t, "hint_output", ai.Hint())
}

func TestAccordionInteractorActionLog(t *testing.T) {
	ag := newMockAccordionGame()
	ap := newMockAccordionPresenter()
	ai := NewAccordionInteractor(ag, ap)

	ap.On("ActionLogOutput", ag).Return("log_output")

	assert.Equal(t, "log_output", ai.ActionLog())
}

func TestAccordionInteractorUndo(t *testing.T) {
	ag := newMockAccordionGame()
	ap := newMockAccordionPresenter()
	ai := NewAccordionInteractor(ag, ap)

	ag.On("Undo").Return(nil)
	ap.On("Output", ag, nil).Return("undo_output")

	assert.Equal(t, "undo_output", ai.Undo())
}

func TestAccordionInteractorUndoN(t *testing.T) {
	ag := newMockAccordionGame()
	ap := newMockAccordionPresenter()
	ai := NewAccordionInteractor(ag, ap)

	ag.On("UndoN", 3).Return(nil)
	ap.On("Output", ag, nil).Return("undo_n_output")

	assert.Equal(t, "undo_n_output", ai.UndoN(3))
}

func TestRestoreAccordionInteractor(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{"tc":{},"pl":[],"ps":0,"mc":0,"al":[],"sl":false}`)
		ap := newMockAccordionPresenter()
		ai, err := RestoreAccordionInteractor(data, ap)
		assert.NoError(t, err)
		assert.NotNil(t, ai)
	})

	t.Run("invalid data", func(t *testing.T) {
		ap := newMockAccordionPresenter()
		_, err := RestoreAccordionInteractor([]byte("invalid"), ap)
		assert.Error(t, err)
	})
}

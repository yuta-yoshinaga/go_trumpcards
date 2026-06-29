package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockSultanGame() *interfaces.MockSultanGame {
	return new(interfaces.MockSultanGame)
}

func newMockSultanPresenter() *presenter.MockSultanPresenter {
	return new(presenter.MockSultanPresenter)
}

func TestNewSultanInteractor(t *testing.T) {
	sg := newMockSultanGame()
	sp := newMockSultanPresenter()
	si := NewSultanInteractor(sg, sp)
	assert.NotNil(t, si)
}

func TestNewSultanInteractorPanicsOnNil(t *testing.T) {
	sp := newMockSultanPresenter()
	assert.Panics(t, func() { NewSultanInteractor(nil, sp) })
	sg := newMockSultanGame()
	assert.Panics(t, func() { NewSultanInteractor(sg, nil) })
}

func TestSultanInteractorReset(t *testing.T) {
	sg := newMockSultanGame()
	sp := newMockSultanPresenter()
	si := NewSultanInteractor(sg, sp)

	sg.On("Reset").Return()
	sp.On("Output", sg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", si.Reset())
	sg.AssertCalled(t, "Reset")
}

func TestSultanInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		sg.On("Draw").Return(nil)
		sp.On("Output", sg, nil).Return("draw_output")
		assert.Equal(t, "draw_output", si.Draw())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		err := errors.New("no cards")
		sg.On("Draw").Return(err)
		sp.On("Output", sg, err).Return("error_output")
		assert.Equal(t, "error_output", si.Draw())
	})
}

func TestSultanInteractorRedeal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		sg.On("Redeal").Return(nil)
		sp.On("Output", sg, nil).Return("redeal_output")
		assert.Equal(t, "redeal_output", si.Redeal())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		err := errors.New("no redeals left")
		sg.On("Redeal").Return(err)
		sp.On("Output", sg, err).Return("error_output")
		assert.Equal(t, "error_output", si.Redeal())
	})
}

func TestSultanInteractorMoveDivanToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		sg.On("MoveDivanToFoundation", 3).Return(nil)
		sp.On("Output", sg, nil).Return("move_output")
		assert.Equal(t, "move_output", si.MoveDivanToFoundation(3))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		err := errors.New("cannot place")
		sg.On("MoveDivanToFoundation", 3).Return(err)
		sp.On("Output", sg, err).Return("error_output")
		assert.Equal(t, "error_output", si.MoveDivanToFoundation(3))
	})
}

func TestSultanInteractorMoveWasteToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		sg.On("MoveWasteToFoundation").Return(nil)
		sp.On("Output", sg, nil).Return("move_output")
		assert.Equal(t, "move_output", si.MoveWasteToFoundation())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		err := errors.New("cannot place")
		sg.On("MoveWasteToFoundation").Return(err)
		sp.On("Output", sg, err).Return("error_output")
		assert.Equal(t, "error_output", si.MoveWasteToFoundation())
	})
}

func TestSultanInteractorGiveUp(t *testing.T) {
	sg := newMockSultanGame()
	sp := newMockSultanPresenter()
	si := NewSultanInteractor(sg, sp)
	sg.On("GiveUp").Return()
	sp.On("Output", sg, nil).Return("giveup_output")
	assert.Equal(t, "giveup_output", si.GiveUp())
	sg.AssertCalled(t, "GiveUp")
}

func TestSultanInteractorHint(t *testing.T) {
	sg := newMockSultanGame()
	sp := newMockSultanPresenter()
	si := NewSultanInteractor(sg, sp)
	sp.On("HintOutput", mock.Anything).Return("hint_output")
	assert.Equal(t, "hint_output", si.Hint())
}

func TestSultanInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		sg.On("AutoComplete").Return(nil)
		sp.On("Output", sg, nil).Return("ac_output")
		assert.Equal(t, "ac_output", si.AutoComplete())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		err := errors.New("not all face up")
		sg.On("AutoComplete").Return(err)
		sp.On("Output", sg, err).Return("error_output")
		assert.Equal(t, "error_output", si.AutoComplete())
	})
}

func TestSultanInteractorActionLog(t *testing.T) {
	sg := newMockSultanGame()
	sp := newMockSultanPresenter()
	si := NewSultanInteractor(sg, sp)
	sp.On("ActionLogOutput", mock.Anything).Return("log_output")
	assert.Equal(t, "log_output", si.ActionLog())
}

func TestSultanInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		sg.On("Undo").Return(nil)
		sp.On("Output", sg, nil).Return("undo_output")
		assert.Equal(t, "undo_output", si.Undo())
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		err := errors.New("nothing to undo")
		sg.On("Undo").Return(err)
		sp.On("Output", sg, err).Return("error_output")
		assert.Equal(t, "error_output", si.Undo())
	})
}

func TestSultanInteractorUndoN(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		sg.On("UndoN", 3).Return(nil)
		sp.On("Output", sg, nil).Return("undon_output")
		assert.Equal(t, "undon_output", si.UndoN(3))
	})
	t.Run("error", func(t *testing.T) {
		sg := newMockSultanGame()
		sp := newMockSultanPresenter()
		si := NewSultanInteractor(sg, sp)
		err := errors.New("undo failed")
		sg.On("UndoN", 3).Return(err)
		sp.On("Output", sg, err).Return("error_output")
		assert.Equal(t, "error_output", si.UndoN(3))
	})
}

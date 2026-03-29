//go:build test

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockTriPeaksGame() *interfaces.MockTriPeaksGame {
	return new(interfaces.MockTriPeaksGame)
}

func newMockTriPeaksPresenter() *presenter.MockTriPeaksPresenter {
	return new(presenter.MockTriPeaksPresenter)
}

func TestNewTriPeaksInteractor(t *testing.T) {
	tg := newMockTriPeaksGame()
	tp := newMockTriPeaksPresenter()
	ti := NewTriPeaksInteractor(tg, tp)
	assert.NotNil(t, ti)
}

func TestNewTriPeaksInteractorPanicsOnNil(t *testing.T) {
	tp := newMockTriPeaksPresenter()
	assert.Panics(t, func() { NewTriPeaksInteractor(nil, tp) })
	tg := newMockTriPeaksGame()
	assert.Panics(t, func() { NewTriPeaksInteractor(tg, nil) })
}

func TestTriPeaksInteractorReset(t *testing.T) {
	tg := newMockTriPeaksGame()
	tp := newMockTriPeaksPresenter()
	ti := NewTriPeaksInteractor(tg, tp)

	tg.On("Reset").Return()
	tp.On("Output", tg, nil).Return("reset_output")

	result := ti.Reset()
	assert.Equal(t, "reset_output", result)
	tg.AssertCalled(t, "Reset")
}

func TestTriPeaksInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tg := newMockTriPeaksGame()
		tp := newMockTriPeaksPresenter()
		ti := NewTriPeaksInteractor(tg, tp)

		tg.On("Draw").Return(nil)
		tp.On("Output", tg, nil).Return("draw_output")

		result := ti.Draw()
		assert.Equal(t, "draw_output", result)
	})

	t.Run("error", func(t *testing.T) {
		tg := newMockTriPeaksGame()
		tp := newMockTriPeaksPresenter()
		ti := NewTriPeaksInteractor(tg, tp)

		drawErr := errors.New("no cards in stock")
		tg.On("Draw").Return(drawErr)
		tp.On("Output", tg, drawErr).Return("draw_error")

		result := ti.Draw()
		assert.Equal(t, "draw_error", result)
	})
}

func TestTriPeaksInteractorRemove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tg := newMockTriPeaksGame()
		tp := newMockTriPeaksPresenter()
		ti := NewTriPeaksInteractor(tg, tp)

		tg.On("Remove", 3, 0).Return(nil)
		tp.On("Output", tg, nil).Return("remove_output")

		result := ti.Remove(3, 0)
		assert.Equal(t, "remove_output", result)
	})

	t.Run("error", func(t *testing.T) {
		tg := newMockTriPeaksGame()
		tp := newMockTriPeaksPresenter()
		ti := NewTriPeaksInteractor(tg, tp)

		removeErr := errors.New("card is not adjacent")
		tg.On("Remove", 3, 0).Return(removeErr)
		tp.On("Output", tg, removeErr).Return("remove_error")

		result := ti.Remove(3, 0)
		assert.Equal(t, "remove_error", result)
	})
}

func TestTriPeaksInteractorGiveUp(t *testing.T) {
	tg := newMockTriPeaksGame()
	tp := newMockTriPeaksPresenter()
	ti := NewTriPeaksInteractor(tg, tp)

	tg.On("GiveUp").Return()
	tp.On("Output", tg, nil).Return("giveup_output")

	result := ti.GiveUp()
	assert.Equal(t, "giveup_output", result)
}

func TestTriPeaksInteractorHint(t *testing.T) {
	tg := newMockTriPeaksGame()
	tp := newMockTriPeaksPresenter()
	ti := NewTriPeaksInteractor(tg, tp)

	tp.On("HintOutput", tg).Return("hint_output")

	result := ti.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestTriPeaksInteractorActionLog(t *testing.T) {
	tg := newMockTriPeaksGame()
	tp := newMockTriPeaksPresenter()
	ti := NewTriPeaksInteractor(tg, tp)

	tp.On("ActionLogOutput", tg).Return("log_output")

	result := ti.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestTriPeaksInteractorUndo(t *testing.T) {
	tg := newMockTriPeaksGame()
	tp := newMockTriPeaksPresenter()
	ti := NewTriPeaksInteractor(tg, tp)

	tg.On("Undo").Return(nil)
	tp.On("Output", tg, nil).Return("undo_output")

	result := ti.Undo()
	assert.Equal(t, "undo_output", result)
}

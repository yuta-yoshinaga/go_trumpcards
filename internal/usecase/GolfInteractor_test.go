//go:build test

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockGolfGame() *interfaces.MockGolfGame {
	return new(interfaces.MockGolfGame)
}

func newMockGolfPresenter() *presenter.MockGolfPresenter {
	return new(presenter.MockGolfPresenter)
}

func TestNewGolfInteractor(t *testing.T) {
	gg := newMockGolfGame()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(gg, gp)
	assert.NotNil(t, gi)
}

func TestNewGolfInteractorPanicsOnNil(t *testing.T) {
	gp := newMockGolfPresenter()
	assert.Panics(t, func() { NewGolfInteractor(nil, gp) })
	gg := newMockGolfGame()
	assert.Panics(t, func() { NewGolfInteractor(gg, nil) })
}

func TestGolfInteractorReset(t *testing.T) {
	gg := newMockGolfGame()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(gg, gp)

	gg.On("Reset").Return()
	gp.On("Output", gg, nil).Return("reset_output")

	result := gi.Reset()
	assert.Equal(t, "reset_output", result)
	gg.AssertCalled(t, "Reset")
}

func TestGolfInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gg := newMockGolfGame()
		gp := newMockGolfPresenter()
		gi := NewGolfInteractor(gg, gp)

		gg.On("Draw").Return(nil)
		gp.On("Output", gg, nil).Return("draw_output")

		result := gi.Draw()
		assert.Equal(t, "draw_output", result)
	})

	t.Run("error", func(t *testing.T) {
		gg := newMockGolfGame()
		gp := newMockGolfPresenter()
		gi := NewGolfInteractor(gg, gp)

		drawErr := errors.New("no cards in stock")
		gg.On("Draw").Return(drawErr)
		gp.On("Output", gg, drawErr).Return("draw_error")

		result := gi.Draw()
		assert.Equal(t, "draw_error", result)
	})
}

func TestGolfInteractorRemove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		gg := newMockGolfGame()
		gp := newMockGolfPresenter()
		gi := NewGolfInteractor(gg, gp)

		gg.On("Remove", 3).Return(nil)
		gp.On("Output", gg, nil).Return("remove_output")

		result := gi.Remove(3)
		assert.Equal(t, "remove_output", result)
	})

	t.Run("error", func(t *testing.T) {
		gg := newMockGolfGame()
		gp := newMockGolfPresenter()
		gi := NewGolfInteractor(gg, gp)

		removeErr := errors.New("card is not adjacent")
		gg.On("Remove", 3).Return(removeErr)
		gp.On("Output", gg, removeErr).Return("remove_error")

		result := gi.Remove(3)
		assert.Equal(t, "remove_error", result)
	})
}

func TestGolfInteractorGiveUp(t *testing.T) {
	gg := newMockGolfGame()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(gg, gp)

	gg.On("GiveUp").Return()
	gp.On("Output", gg, nil).Return("giveup_output")

	result := gi.GiveUp()
	assert.Equal(t, "giveup_output", result)
}

func TestGolfInteractorHint(t *testing.T) {
	gg := newMockGolfGame()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(gg, gp)

	gp.On("HintOutput", gg).Return("hint_output")

	result := gi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestGolfInteractorActionLog(t *testing.T) {
	gg := newMockGolfGame()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(gg, gp)

	gp.On("ActionLogOutput", gg).Return("log_output")

	result := gi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestGolfInteractorUndo(t *testing.T) {
	gg := newMockGolfGame()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(gg, gp)

	gg.On("Undo").Return(nil)
	gp.On("Output", gg, nil).Return("undo_output")

	result := gi.Undo()
	assert.Equal(t, "undo_output", result)
}

func TestGolfInteractorSnapshot(t *testing.T) {
	game := domain.NewGolf(domain.NewTrumpCards(0))
	game.Reset()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(game, gp)

	data, err := gi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreGolfInteractor(t *testing.T) {
	game := domain.NewGolf(domain.NewTrumpCards(0))
	game.Reset()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(game, gp)

	data, err := gi.Snapshot()
	assert.NoError(t, err)

	restored, err := RestoreGolfInteractor(data, gp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreGolfInteractorInvalidJSON(t *testing.T) {
	gp := newMockGolfPresenter()
	_, err := RestoreGolfInteractor([]byte("invalid"), gp)
	assert.Error(t, err)
}

// r9 は CUI 側の集計をプレゼンタが持っているので、インタラクタはドメインを
// 触らずプレゼンタへ委譲する。
func TestGolfInteractorResetNineHole(t *testing.T) {
	gg := newMockGolfGame()
	gp := newMockGolfPresenter()
	gi := NewGolfInteractor(gg, gp)

	gp.On("ResetNineHole", gg).Return("reset9_output")

	assert.Equal(t, "reset9_output", gi.ResetNineHole())
	gp.AssertExpectations(t)
}

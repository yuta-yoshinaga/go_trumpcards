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

func newMockAcesUpGame() *interfaces.MockAcesUpGame {
	return new(interfaces.MockAcesUpGame)
}

func newMockAcesUpPresenter() *presenter.MockAcesUpPresenter {
	return new(presenter.MockAcesUpPresenter)
}

func TestNewAcesUpInteractor(t *testing.T) {
	g := newMockAcesUpGame()
	gp := newMockAcesUpPresenter()
	assert.NotNil(t, NewAcesUpInteractor(g, gp))
}

func TestNewAcesUpInteractorPanicsOnNil(t *testing.T) {
	gp := newMockAcesUpPresenter()
	assert.Panics(t, func() { NewAcesUpInteractor(nil, gp) })
	g := newMockAcesUpGame()
	assert.Panics(t, func() { NewAcesUpInteractor(g, nil) })
}

func TestAcesUpInteractorReset(t *testing.T) {
	g := newMockAcesUpGame()
	gp := newMockAcesUpPresenter()
	ai := NewAcesUpInteractor(g, gp)

	g.On("Reset").Return()
	gp.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", ai.Reset())
	g.AssertCalled(t, "Reset")
}

func TestAcesUpInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockAcesUpGame()
		gp := newMockAcesUpPresenter()
		ai := NewAcesUpInteractor(g, gp)
		g.On("Draw").Return(nil)
		gp.On("Output", g, nil).Return("draw_output")
		assert.Equal(t, "draw_output", ai.Draw())
	})
	t.Run("error", func(t *testing.T) {
		g := newMockAcesUpGame()
		gp := newMockAcesUpPresenter()
		ai := NewAcesUpInteractor(g, gp)
		drawErr := errors.New("no cards in stock")
		g.On("Draw").Return(drawErr)
		gp.On("Output", g, drawErr).Return("draw_error")
		assert.Equal(t, "draw_error", ai.Draw())
	})
}

func TestAcesUpInteractorRemove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockAcesUpGame()
		gp := newMockAcesUpPresenter()
		ai := NewAcesUpInteractor(g, gp)
		g.On("Remove", 2).Return(nil)
		gp.On("Output", g, nil).Return("remove_output")
		assert.Equal(t, "remove_output", ai.Remove(2))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockAcesUpGame()
		gp := newMockAcesUpPresenter()
		ai := NewAcesUpInteractor(g, gp)
		removeErr := errors.New("card is not removable")
		g.On("Remove", 2).Return(removeErr)
		gp.On("Output", g, removeErr).Return("remove_error")
		assert.Equal(t, "remove_error", ai.Remove(2))
	})
}

func TestAcesUpInteractorMove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockAcesUpGame()
		gp := newMockAcesUpPresenter()
		ai := NewAcesUpInteractor(g, gp)
		g.On("Move", 1).Return(nil)
		gp.On("Output", g, nil).Return("move_output")
		assert.Equal(t, "move_output", ai.Move(1))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockAcesUpGame()
		gp := newMockAcesUpPresenter()
		ai := NewAcesUpInteractor(g, gp)
		moveErr := errors.New("no empty column")
		g.On("Move", 1).Return(moveErr)
		gp.On("Output", g, moveErr).Return("move_error")
		assert.Equal(t, "move_error", ai.Move(1))
	})
}

func TestAcesUpInteractorGiveUp(t *testing.T) {
	g := newMockAcesUpGame()
	gp := newMockAcesUpPresenter()
	ai := NewAcesUpInteractor(g, gp)
	g.On("GiveUp").Return()
	gp.On("Output", g, nil).Return("giveup_output")
	assert.Equal(t, "giveup_output", ai.GiveUp())
}

func TestAcesUpInteractorHint(t *testing.T) {
	g := newMockAcesUpGame()
	gp := newMockAcesUpPresenter()
	ai := NewAcesUpInteractor(g, gp)
	gp.On("HintOutput", g).Return("hint_output")
	assert.Equal(t, "hint_output", ai.Hint())
}

func TestAcesUpInteractorActionLog(t *testing.T) {
	g := newMockAcesUpGame()
	gp := newMockAcesUpPresenter()
	ai := NewAcesUpInteractor(g, gp)
	gp.On("ActionLogOutput", g).Return("log_output")
	assert.Equal(t, "log_output", ai.ActionLog())
}

func TestAcesUpInteractorUndo(t *testing.T) {
	g := newMockAcesUpGame()
	gp := newMockAcesUpPresenter()
	ai := NewAcesUpInteractor(g, gp)
	g.On("Undo").Return(nil)
	gp.On("Output", g, nil).Return("undo_output")
	assert.Equal(t, "undo_output", ai.Undo())
}

func TestAcesUpInteractorUndoN(t *testing.T) {
	g := newMockAcesUpGame()
	gp := newMockAcesUpPresenter()
	ai := NewAcesUpInteractor(g, gp)
	g.On("UndoN", 3).Return(nil)
	gp.On("Output", g, nil).Return("undon_output")
	assert.Equal(t, "undon_output", ai.UndoN(3))
}

func TestAcesUpInteractorSnapshot(t *testing.T) {
	game := domain.NewAcesUp(domain.NewTrumpCards(0))
	game.Reset()
	gp := newMockAcesUpPresenter()
	ai := NewAcesUpInteractor(game, gp)

	data, err := ai.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreAcesUpInteractor(t *testing.T) {
	game := domain.NewAcesUp(domain.NewTrumpCards(0))
	game.Reset()
	gp := newMockAcesUpPresenter()
	ai := NewAcesUpInteractor(game, gp)

	data, err := ai.Snapshot()
	assert.NoError(t, err)

	restored, err := RestoreAcesUpInteractor(data, gp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreAcesUpInteractorInvalidJSON(t *testing.T) {
	gp := newMockAcesUpPresenter()
	_, err := RestoreAcesUpInteractor([]byte("invalid"), gp)
	assert.Error(t, err)
}

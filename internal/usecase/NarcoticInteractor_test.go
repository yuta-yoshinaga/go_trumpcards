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

func newMockNarcoticGame() *interfaces.MockNarcoticGame {
	return new(interfaces.MockNarcoticGame)
}

func newMockNarcoticPresenter() *presenter.MockNarcoticPresenter {
	return new(presenter.MockNarcoticPresenter)
}

func TestNewNarcoticInteractor(t *testing.T) {
	g := newMockNarcoticGame()
	gp := newMockNarcoticPresenter()
	assert.NotNil(t, NewNarcoticInteractor(g, gp))
}

func TestNewNarcoticInteractorPanicsOnNil(t *testing.T) {
	gp := newMockNarcoticPresenter()
	assert.Panics(t, func() { NewNarcoticInteractor(nil, gp) })
	g := newMockNarcoticGame()
	assert.Panics(t, func() { NewNarcoticInteractor(g, nil) })
}

func TestNarcoticInteractorReset(t *testing.T) {
	g := newMockNarcoticGame()
	gp := newMockNarcoticPresenter()
	ai := NewNarcoticInteractor(g, gp)

	g.On("Reset").Return()
	gp.On("Output", g, nil).Return("reset_output")

	assert.Equal(t, "reset_output", ai.Reset())
	g.AssertCalled(t, "Reset")
}

func TestNarcoticInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockNarcoticGame()
		gp := newMockNarcoticPresenter()
		ai := NewNarcoticInteractor(g, gp)
		g.On("Draw").Return(nil)
		gp.On("Output", g, nil).Return("draw_output")
		assert.Equal(t, "draw_output", ai.Draw())
	})
	t.Run("error", func(t *testing.T) {
		g := newMockNarcoticGame()
		gp := newMockNarcoticPresenter()
		ai := NewNarcoticInteractor(g, gp)
		drawErr := errors.New("no cards in stock")
		g.On("Draw").Return(drawErr)
		gp.On("Output", g, drawErr).Return("draw_error")
		assert.Equal(t, "draw_error", ai.Draw())
	})
}

func TestNarcoticInteractorRemove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockNarcoticGame()
		gp := newMockNarcoticPresenter()
		ai := NewNarcoticInteractor(g, gp)
		g.On("Remove").Return(nil)
		gp.On("Output", g, nil).Return("remove_output")
		assert.Equal(t, "remove_output", ai.Remove())
	})
	t.Run("error", func(t *testing.T) {
		g := newMockNarcoticGame()
		gp := newMockNarcoticPresenter()
		ai := NewNarcoticInteractor(g, gp)
		removeErr := errors.New("card is not removable")
		g.On("Remove").Return(removeErr)
		gp.On("Output", g, removeErr).Return("remove_error")
		assert.Equal(t, "remove_error", ai.Remove())
	})
}

func TestNarcoticInteractorMove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockNarcoticGame()
		gp := newMockNarcoticPresenter()
		ai := NewNarcoticInteractor(g, gp)
		g.On("Move", 1).Return(nil)
		gp.On("Output", g, nil).Return("move_output")
		assert.Equal(t, "move_output", ai.Move(1))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockNarcoticGame()
		gp := newMockNarcoticPresenter()
		ai := NewNarcoticInteractor(g, gp)
		moveErr := errors.New("no empty column")
		g.On("Move", 1).Return(moveErr)
		gp.On("Output", g, moveErr).Return("move_error")
		assert.Equal(t, "move_error", ai.Move(1))
	})
}

func TestNarcoticInteractorGiveUp(t *testing.T) {
	g := newMockNarcoticGame()
	gp := newMockNarcoticPresenter()
	ai := NewNarcoticInteractor(g, gp)
	g.On("GiveUp").Return()
	gp.On("Output", g, nil).Return("giveup_output")
	assert.Equal(t, "giveup_output", ai.GiveUp())
}

func TestNarcoticInteractorHint(t *testing.T) {
	g := newMockNarcoticGame()
	gp := newMockNarcoticPresenter()
	ai := NewNarcoticInteractor(g, gp)
	gp.On("HintOutput", g).Return("hint_output")
	assert.Equal(t, "hint_output", ai.Hint())
}

func TestNarcoticInteractorActionLog(t *testing.T) {
	g := newMockNarcoticGame()
	gp := newMockNarcoticPresenter()
	ai := NewNarcoticInteractor(g, gp)
	gp.On("ActionLogOutput", g).Return("log_output")
	assert.Equal(t, "log_output", ai.ActionLog())
}

func TestNarcoticInteractorUndo(t *testing.T) {
	g := newMockNarcoticGame()
	gp := newMockNarcoticPresenter()
	ai := NewNarcoticInteractor(g, gp)
	g.On("Undo").Return(nil)
	gp.On("Output", g, nil).Return("undo_output")
	assert.Equal(t, "undo_output", ai.Undo())
}

func TestNarcoticInteractorUndoN(t *testing.T) {
	g := newMockNarcoticGame()
	gp := newMockNarcoticPresenter()
	ai := NewNarcoticInteractor(g, gp)
	g.On("UndoN", 3).Return(nil)
	gp.On("Output", g, nil).Return("undon_output")
	assert.Equal(t, "undon_output", ai.UndoN(3))
}

func TestNarcoticInteractorSnapshot(t *testing.T) {
	game := domain.NewNarcotic(domain.NewTrumpCards(0))
	game.Reset()
	gp := newMockNarcoticPresenter()
	ai := NewNarcoticInteractor(game, gp)

	data, err := ai.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreNarcoticInteractor(t *testing.T) {
	game := domain.NewNarcotic(domain.NewTrumpCards(0))
	game.Reset()
	gp := newMockNarcoticPresenter()
	ai := NewNarcoticInteractor(game, gp)

	data, err := ai.Snapshot()
	assert.NoError(t, err)

	restored, err := RestoreNarcoticInteractor(data, gp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreNarcoticInteractorInvalidJSON(t *testing.T) {
	gp := newMockNarcoticPresenter()
	_, err := RestoreNarcoticInteractor([]byte("invalid"), gp)
	assert.Error(t, err)
}

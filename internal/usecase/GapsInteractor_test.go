package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockGapsGame() *interfaces.MockGapsGame {
	return new(interfaces.MockGapsGame)
}

func newMockGapsPresenter() *presenter.MockGapsPresenter {
	return new(presenter.MockGapsPresenter)
}

func TestNewGapsInteractor(t *testing.T) {
	g := newMockGapsGame()
	gp := newMockGapsPresenter()
	gi := NewGapsInteractor(g, gp)
	assert.NotNil(t, gi)
}

func TestNewGapsInteractor_PanicsOnNil(t *testing.T) {
	gp := newMockGapsPresenter()
	assert.Panics(t, func() { NewGapsInteractor(nil, gp) })
	g := newMockGapsGame()
	assert.Panics(t, func() { NewGapsInteractor(g, nil) })
}

func TestGapsInteractor_Reset(t *testing.T) {
	g := newMockGapsGame()
	gp := newMockGapsPresenter()
	gi := NewGapsInteractor(g, gp)
	g.On("Reset").Return()
	gp.On("Output", g, nil).Return("reset_out")
	assert.Equal(t, "reset_out", gi.Reset())
	g.AssertCalled(t, "Reset")
}

func TestGapsInteractor_Move(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockGapsGame()
		gp := newMockGapsPresenter()
		gi := NewGapsInteractor(g, gp)
		g.On("Move", 0, 0, 1, 1).Return(nil)
		gp.On("Output", g, nil).Return("move_out")
		assert.Equal(t, "move_out", gi.Move(0, 0, 1, 1))
	})
	t.Run("error", func(t *testing.T) {
		g := newMockGapsGame()
		gp := newMockGapsPresenter()
		gi := NewGapsInteractor(g, gp)
		err := errors.New("illegal")
		g.On("Move", 0, 0, 1, 1).Return(err)
		gp.On("Output", g, err).Return("move_err")
		assert.Equal(t, "move_err", gi.Move(0, 0, 1, 1))
	})
}

func TestGapsInteractor_Redeal(t *testing.T) {
	g := newMockGapsGame()
	gp := newMockGapsPresenter()
	gi := NewGapsInteractor(g, gp)
	g.On("Redeal").Return(nil)
	gp.On("Output", g, nil).Return("redeal_out")
	assert.Equal(t, "redeal_out", gi.Redeal())
}

func TestGapsInteractor_Undo(t *testing.T) {
	g := newMockGapsGame()
	gp := newMockGapsPresenter()
	gi := NewGapsInteractor(g, gp)
	g.On("Undo").Return(nil)
	gp.On("Output", g, nil).Return("undo_out")
	assert.Equal(t, "undo_out", gi.Undo())
}

func TestGapsInteractor_UndoN(t *testing.T) {
	g := newMockGapsGame()
	gp := newMockGapsPresenter()
	gi := NewGapsInteractor(g, gp)
	g.On("UndoN", 3).Return(nil)
	gp.On("Output", g, nil).Return("undon_out")
	assert.Equal(t, "undon_out", gi.UndoN(3))
}

func TestGapsInteractor_GiveUp(t *testing.T) {
	g := newMockGapsGame()
	gp := newMockGapsPresenter()
	gi := NewGapsInteractor(g, gp)
	g.On("GiveUp").Return()
	gp.On("Output", g, nil).Return("giveup_out")
	assert.Equal(t, "giveup_out", gi.GiveUp())
}

func TestGapsInteractor_Hint(t *testing.T) {
	g := newMockGapsGame()
	gp := newMockGapsPresenter()
	gi := NewGapsInteractor(g, gp)
	gp.On("HintOutput", g).Return("hint_out")
	assert.Equal(t, "hint_out", gi.Hint())
}

func TestGapsInteractor_ActionLog(t *testing.T) {
	g := newMockGapsGame()
	gp := newMockGapsPresenter()
	gi := NewGapsInteractor(g, gp)
	gp.On("ActionLogOutput", g).Return("log_out")
	assert.Equal(t, "log_out", gi.ActionLog())
}

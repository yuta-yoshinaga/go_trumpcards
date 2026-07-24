package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockClockSolitaireGame() *interfaces.MockClockSolitaireGame {
	return new(interfaces.MockClockSolitaireGame)
}

func newMockClockSolitairePresenter() *presenter.MockClockSolitairePresenter {
	return new(presenter.MockClockSolitairePresenter)
}

func TestNewClockSolitaireInteractor(t *testing.T) {
	g := newMockClockSolitaireGame()
	p := newMockClockSolitairePresenter()
	ci := NewClockSolitaireInteractor(g, p)
	assert.NotNil(t, ci)
}

func TestNewClockSolitaireInteractorPanicsOnNil(t *testing.T) {
	p := newMockClockSolitairePresenter()
	assert.Panics(t, func() { NewClockSolitaireInteractor(nil, p) })
	g := newMockClockSolitaireGame()
	assert.Panics(t, func() { NewClockSolitaireInteractor(g, nil) })
}

func TestClockSolitaireInteractorReset(t *testing.T) {
	g := newMockClockSolitaireGame()
	p := newMockClockSolitairePresenter()
	ci := NewClockSolitaireInteractor(g, p)

	g.On("Reset").Return()
	p.On("Output", g, nil).Return("reset_output")

	result := ci.Reset()
	assert.Equal(t, "reset_output", result)
	g.AssertCalled(t, "Reset")
}

func TestClockSolitaireInteractorStep(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockClockSolitaireGame()
		p := newMockClockSolitairePresenter()
		ci := NewClockSolitaireInteractor(g, p)

		g.On("Step").Return(nil)
		p.On("Output", g, nil).Return("step_output")

		result := ci.Step()
		assert.Equal(t, "step_output", result)
	})

	t.Run("error", func(t *testing.T) {
		g := newMockClockSolitaireGame()
		p := newMockClockSolitairePresenter()
		ci := NewClockSolitaireInteractor(g, p)

		err := errors.New("not playing")
		g.On("Step").Return(err)
		p.On("Output", g, err).Return("error_output")

		result := ci.Step()
		assert.Equal(t, "error_output", result)
	})
}

func TestClockSolitaireInteractorAutoPlay(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockClockSolitaireGame()
		p := newMockClockSolitairePresenter()
		ci := NewClockSolitaireInteractor(g, p)

		g.On("AutoPlay").Return(nil)
		p.On("Output", g, nil).Return("autoplay_output")

		result := ci.AutoPlay()
		assert.Equal(t, "autoplay_output", result)
	})

	t.Run("error", func(t *testing.T) {
		g := newMockClockSolitaireGame()
		p := newMockClockSolitairePresenter()
		ci := NewClockSolitaireInteractor(g, p)

		err := errors.New("not playing")
		g.On("AutoPlay").Return(err)
		p.On("Output", g, err).Return("error_output")

		result := ci.AutoPlay()
		assert.Equal(t, "error_output", result)
	})
}

func TestClockSolitaireInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		g := newMockClockSolitaireGame()
		p := newMockClockSolitairePresenter()
		ci := NewClockSolitaireInteractor(g, p)

		g.On("Undo").Return(nil)
		p.On("Output", g, nil).Return("undo_output")

		result := ci.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		g := newMockClockSolitaireGame()
		p := newMockClockSolitairePresenter()
		ci := NewClockSolitaireInteractor(g, p)

		err := errors.New("no history")
		g.On("Undo").Return(err)
		p.On("Output", g, err).Return("error_output")

		result := ci.Undo()
		assert.Equal(t, "error_output", result)
	})
}

func TestClockSolitaireInteractorActionLog(t *testing.T) {
	g := newMockClockSolitaireGame()
	p := newMockClockSolitairePresenter()
	ci := NewClockSolitaireInteractor(g, p)

	p.On("ActionLogOutput", g).Return("log_output")

	result := ci.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestClockSolitaireInteractorSnapshot(t *testing.T) {
	cs := newTestClockSolitaireForInteractor()
	p := newMockClockSolitairePresenter()
	ci := NewClockSolitaireInteractor(cs, p)

	data, err := ci.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreClockSolitaireInteractor(t *testing.T) {
	cs := newTestClockSolitaireForInteractor()
	p := newMockClockSolitairePresenter()
	ci := NewClockSolitaireInteractor(cs, p)

	data, err := ci.Snapshot()
	assert.NoError(t, err)

	ci2, err := RestoreClockSolitaireInteractor(data, p)
	assert.NoError(t, err)
	assert.NotNil(t, ci2)
}

func TestRestoreClockSolitaireInteractor_InvalidJSON(t *testing.T) {
	p := newMockClockSolitairePresenter()
	_, err := RestoreClockSolitaireInteractor([]byte("invalid"), p)
	assert.Error(t, err)
}

func newTestClockSolitaireForInteractor() *domain.ClockSolitaire {
	tc := domain.NewTrumpCards(0)
	cs := domain.NewClockSolitaire(tc)
	cs.Reset()
	return cs
}

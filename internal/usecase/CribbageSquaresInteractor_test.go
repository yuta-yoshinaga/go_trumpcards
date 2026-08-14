package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockCribbageSquaresGame() *interfaces.MockCribbageSquaresGame {
	return new(interfaces.MockCribbageSquaresGame)
}

func newMockCribbageSquaresPresenter() *presenter.MockCribbageSquaresPresenter {
	return new(presenter.MockCribbageSquaresPresenter)
}

func TestNewCribbageSquaresInteractor(t *testing.T) {
	pg := newMockCribbageSquaresGame()
	pp := newMockCribbageSquaresPresenter()
	pi := NewCribbageSquaresInteractor(pg, pp)
	assert.NotNil(t, pi)
}

func TestNewCribbageSquaresInteractorPanicsOnNil(t *testing.T) {
	pp := newMockCribbageSquaresPresenter()
	assert.Panics(t, func() { NewCribbageSquaresInteractor(nil, pp) })
	pg := newMockCribbageSquaresGame()
	assert.Panics(t, func() { NewCribbageSquaresInteractor(pg, nil) })
}

func TestCribbageSquaresInteractor_Reset(t *testing.T) {
	pg := newMockCribbageSquaresGame()
	pp := newMockCribbageSquaresPresenter()
	pi := NewCribbageSquaresInteractor(pg, pp)

	pg.On("Reset").Return()
	pp.On("Output", pg, nil).Return("reset_output")

	assert.Equal(t, "reset_output", pi.Reset())
	pg.AssertCalled(t, "Reset")
}

func TestCribbageSquaresInteractor_Place(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pg := newMockCribbageSquaresGame()
		pp := newMockCribbageSquaresPresenter()
		pi := NewCribbageSquaresInteractor(pg, pp)

		pg.On("Place", 0, 0).Return(nil)
		pp.On("Output", pg, nil).Return("place_output")

		assert.Equal(t, "place_output", pi.Place(0, 0))
	})

	t.Run("error", func(t *testing.T) {
		pg := newMockCribbageSquaresGame()
		pp := newMockCribbageSquaresPresenter()
		pi := NewCribbageSquaresInteractor(pg, pp)

		e := errors.New("occupied")
		pg.On("Place", 1, 1).Return(e)
		pp.On("Output", pg, e).Return("place_err")

		assert.Equal(t, "place_err", pi.Place(1, 1))
	})
}

func TestCribbageSquaresInteractor_Undo(t *testing.T) {
	pg := newMockCribbageSquaresGame()
	pp := newMockCribbageSquaresPresenter()
	pi := NewCribbageSquaresInteractor(pg, pp)

	pg.On("Undo").Return(nil)
	pp.On("Output", pg, nil).Return("undo_output")

	assert.Equal(t, "undo_output", pi.Undo())
}

func TestCribbageSquaresInteractor_GiveUp(t *testing.T) {
	pg := newMockCribbageSquaresGame()
	pp := newMockCribbageSquaresPresenter()
	pi := NewCribbageSquaresInteractor(pg, pp)

	pg.On("GiveUp").Return()
	pp.On("Output", pg, nil).Return("giveup_output")

	assert.Equal(t, "giveup_output", pi.GiveUp())
}

func TestCribbageSquaresInteractor_Hint(t *testing.T) {
	pg := newMockCribbageSquaresGame()
	pp := newMockCribbageSquaresPresenter()
	pi := NewCribbageSquaresInteractor(pg, pp)

	pp.On("HintOutput", pg).Return("hint_output")

	assert.Equal(t, "hint_output", pi.Hint())
	pp.AssertCalled(t, "HintOutput", pg)
}

func TestCribbageSquaresInteractor_ActionLog(t *testing.T) {
	pg := newMockCribbageSquaresGame()
	pp := newMockCribbageSquaresPresenter()
	pi := NewCribbageSquaresInteractor(pg, pp)

	pp.On("ActionLogOutput", pg).Return("log_output")

	assert.Equal(t, "log_output", pi.ActionLog())
}

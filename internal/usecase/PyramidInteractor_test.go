package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockPyramidGame() *interfaces.MockPyramidGame {
	return new(interfaces.MockPyramidGame)
}

func newMockPyramidPresenter() *presenter.MockPyramidPresenter {
	return new(presenter.MockPyramidPresenter)
}

func TestNewPyramidInteractor(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)
	assert.NotNil(t, pi)
}

func TestNewPyramidInteractorPanicsOnNil(t *testing.T) {
	pp := newMockPyramidPresenter()
	assert.Panics(t, func() { NewPyramidInteractor(nil, pp) })
	pg := newMockPyramidGame()
	assert.Panics(t, func() { NewPyramidInteractor(pg, nil) })
}

func TestPyramidInteractorReset(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pg.On("Reset").Return()
	pp.On("Output", pg, nil).Return("reset_output")

	result := pi.Reset()
	assert.Equal(t, "reset_output", result)
	pg.AssertCalled(t, "Reset")
}

func TestPyramidInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pg := newMockPyramidGame()
		pp := newMockPyramidPresenter()
		pi := NewPyramidInteractor(pg, pp)

		pg.On("Draw").Return(nil)
		pp.On("Output", pg, nil).Return("draw_output")

		result := pi.Draw()
		assert.Equal(t, "draw_output", result)
	})

	t.Run("error", func(t *testing.T) {
		pg := newMockPyramidGame()
		pp := newMockPyramidPresenter()
		pi := NewPyramidInteractor(pg, pp)

		err := errors.New("no cards")
		pg.On("Draw").Return(err)
		pp.On("Output", pg, err).Return("error_output")

		result := pi.Draw()
		assert.Equal(t, "error_output", result)
	})
}

func TestPyramidInteractorRemovePair(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pg.On("RemovePair", 6, 0, 6, 1).Return(nil)
	pp.On("Output", pg, nil).Return("remove_pair_output")

	result := pi.RemovePair(6, 0, 6, 1)
	assert.Equal(t, "remove_pair_output", result)
}

func TestPyramidInteractorRemoveKing(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pg.On("RemoveKing", 6, 0).Return(nil)
	pp.On("Output", pg, nil).Return("remove_king_output")

	result := pi.RemoveKing(6, 0)
	assert.Equal(t, "remove_king_output", result)
}

func TestPyramidInteractorRemoveWithWaste(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pg.On("RemoveWithWaste", 6, 0).Return(nil)
	pp.On("Output", pg, nil).Return("remove_waste_output")

	result := pi.RemoveWithWaste(6, 0)
	assert.Equal(t, "remove_waste_output", result)
}

func TestPyramidInteractorRemoveWasteKing(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pg.On("RemoveWasteKing").Return(nil)
	pp.On("Output", pg, nil).Return("remove_wk_output")

	result := pi.RemoveWasteKing()
	assert.Equal(t, "remove_wk_output", result)
}

func TestPyramidInteractorGiveUp(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pg.On("GiveUp").Return()
	pp.On("Output", pg, nil).Return("giveup_output")

	result := pi.GiveUp()
	assert.Equal(t, "giveup_output", result)
}

func TestPyramidInteractorHint(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pp.On("HintOutput", pg).Return("hint_output")

	result := pi.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestPyramidInteractorActionLog(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pp.On("ActionLogOutput", pg).Return("log_output")

	result := pi.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestPyramidInteractorUndo(t *testing.T) {
	pg := newMockPyramidGame()
	pp := newMockPyramidPresenter()
	pi := NewPyramidInteractor(pg, pp)

	pg.On("Undo").Return(nil)
	pp.On("Output", pg, nil).Return("undo_output")

	result := pi.Undo()
	assert.Equal(t, "undo_output", result)
}

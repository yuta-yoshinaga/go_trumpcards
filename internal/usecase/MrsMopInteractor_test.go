package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockMrsMopGame() *interfaces.MockMrsMopGame {
	return new(interfaces.MockMrsMopGame)
}

func newMockMrsMopPresenter() *presenter.MockMrsMopPresenter {
	return new(presenter.MockMrsMopPresenter)
}

func TestNewMrsMopInteractor(t *testing.T) {
	sg := newMockMrsMopGame()
	sp := newMockMrsMopPresenter()
	si := NewMrsMopInteractor(sg, sp)
	assert.NotNil(t, si)
}

func TestNewMrsMopInteractorPanicsOnNil(t *testing.T) {
	sp := newMockMrsMopPresenter()
	assert.Panics(t, func() { NewMrsMopInteractor(nil, sp) })
	sg := newMockMrsMopGame()
	assert.Panics(t, func() { NewMrsMopInteractor(sg, nil) })
}

func TestMrsMopInteractorReset(t *testing.T) {
	sg := newMockMrsMopGame()
	sp := newMockMrsMopPresenter()
	si := NewMrsMopInteractor(sg, sp)

	sg.On("Reset").Return()
	sp.On("Output", sg, nil).Return("reset_output")

	result := si.Reset()
	assert.Equal(t, "reset_output", result)
	sg.AssertCalled(t, "Reset")
}

func TestMrsMopInteractorResetWithConfig(t *testing.T) {
	sg := newMockMrsMopGame()
	sp := newMockMrsMopPresenter()
	si := NewMrsMopInteractor(sg, sp)

	cfg := domain.MrsMopConfig{Difficulty: domain.MrsMopDifficulty2Suit}
	sg.On("ResetWithConfig", cfg).Return()
	sp.On("Output", sg, nil).Return("reset_config_output")

	result := si.ResetWithConfig(cfg)
	assert.Equal(t, "reset_config_output", result)
	sg.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestMrsMopInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockMrsMopGame()
		sp := newMockMrsMopPresenter()
		si := NewMrsMopInteractor(sg, sp)

		sg.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		sp.On("Output", sg, nil).Return("move_output")

		result := si.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		sg := newMockMrsMopGame()
		sp := newMockMrsMopPresenter()
		si := NewMrsMopInteractor(sg, sp)

		err := errors.New("invalid")
		sg.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		sp.On("Output", sg, err).Return("error_output")

		result := si.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "error_output", result)
	})
}

func TestMrsMopInteractorGiveUp(t *testing.T) {
	sg := newMockMrsMopGame()
	sp := newMockMrsMopPresenter()
	si := NewMrsMopInteractor(sg, sp)

	sg.On("GiveUp").Return()
	sp.On("Output", sg, nil).Return("giveup_output")

	result := si.GiveUp()
	assert.Equal(t, "giveup_output", result)
	sg.AssertCalled(t, "GiveUp")
}

func TestMrsMopInteractorHint(t *testing.T) {
	sg := newMockMrsMopGame()
	sp := newMockMrsMopPresenter()
	si := NewMrsMopInteractor(sg, sp)

	sp.On("HintOutput", mock.Anything).Return("hint_output")

	result := si.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestMrsMopInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockMrsMopGame()
		sp := newMockMrsMopPresenter()
		si := NewMrsMopInteractor(sg, sp)

		sg.On("AutoComplete").Return(nil)
		sp.On("Output", sg, nil).Return("ac_output")

		result := si.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		sg := newMockMrsMopGame()
		sp := newMockMrsMopPresenter()
		si := NewMrsMopInteractor(sg, sp)

		err := errors.New("not all face up")
		sg.On("AutoComplete").Return(err)
		sp.On("Output", sg, err).Return("error_output")

		result := si.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestMrsMopInteractorActionLog(t *testing.T) {
	sg := newMockMrsMopGame()
	sp := newMockMrsMopPresenter()
	si := NewMrsMopInteractor(sg, sp)

	sp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := si.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestMrsMopInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sg := newMockMrsMopGame()
		sp := newMockMrsMopPresenter()
		si := NewMrsMopInteractor(sg, sp)

		sg.On("Undo").Return(nil)
		sp.On("Output", sg, nil).Return("undo_output")

		result := si.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		sg := newMockMrsMopGame()
		sp := newMockMrsMopPresenter()
		si := NewMrsMopInteractor(sg, sp)

		err := errors.New("nothing to undo")
		sg.On("Undo").Return(err)
		sp.On("Output", sg, err).Return("error_output")

		result := si.Undo()
		assert.Equal(t, "error_output", result)
	})
}

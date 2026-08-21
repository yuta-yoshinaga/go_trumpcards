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

func newMockWhiteheadGame() *interfaces.MockWhiteheadGame {
	return new(interfaces.MockWhiteheadGame)
}

func newMockWhiteheadPresenter() *presenter.MockWhiteheadPresenter {
	return new(presenter.MockWhiteheadPresenter)
}

func TestNewWhiteheadInteractor(t *testing.T) {
	kg := newMockWhiteheadGame()
	kp := newMockWhiteheadPresenter()
	ki := NewWhiteheadInteractor(kg, kp)
	assert.NotNil(t, ki)
}

func TestNewWhiteheadInteractorPanicsOnNil(t *testing.T) {
	kp := newMockWhiteheadPresenter()
	assert.Panics(t, func() { NewWhiteheadInteractor(nil, kp) })
	kg := newMockWhiteheadGame()
	assert.Panics(t, func() { NewWhiteheadInteractor(kg, nil) })
}

func TestWhiteheadInteractorReset(t *testing.T) {
	kg := newMockWhiteheadGame()
	kp := newMockWhiteheadPresenter()
	ki := NewWhiteheadInteractor(kg, kp)

	kg.On("Reset").Return()
	kp.On("Output", kg, nil).Return("reset_output")

	result := ki.Reset()
	assert.Equal(t, "reset_output", result)
	kg.AssertCalled(t, "Reset")
}

func TestWhiteheadInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		kg.On("Draw").Return(nil)
		kp.On("Output", kg, nil).Return("draw_output")

		result := ki.Draw()
		assert.Equal(t, "draw_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		err := errors.New("no cards")
		kg.On("Draw").Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.Draw()
		assert.Equal(t, "error_output", result)
	})
}

func TestWhiteheadInteractorMoveWasteToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		kg.On("MoveWasteToTableau", 3).Return(nil)
		kp.On("Output", kg, nil).Return("move_output")

		result := ki.MoveWasteToTableau(3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		err := errors.New("cannot place")
		kg.On("MoveWasteToTableau", 3).Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.MoveWasteToTableau(3)
		assert.Equal(t, "error_output", result)
	})
}

func TestWhiteheadInteractorMoveWasteToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		kg.On("MoveWasteToFoundation").Return(nil)
		kp.On("Output", kg, nil).Return("move_output")

		result := ki.MoveWasteToFoundation()
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		err := errors.New("cannot place")
		kg.On("MoveWasteToFoundation").Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.MoveWasteToFoundation()
		assert.Equal(t, "error_output", result)
	})
}

func TestWhiteheadInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		kg.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		kp.On("Output", kg, nil).Return("move_output")

		result := ki.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		err := errors.New("invalid")
		kg.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "error_output", result)
	})
}

func TestWhiteheadInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		kg.On("MoveTableauToFoundation", 2).Return(nil)
		kp.On("Output", kg, nil).Return("move_output")

		result := ki.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		err := errors.New("invalid")
		kg.On("MoveTableauToFoundation", 2).Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestWhiteheadInteractorGiveUp(t *testing.T) {
	kg := newMockWhiteheadGame()
	kp := newMockWhiteheadPresenter()
	ki := NewWhiteheadInteractor(kg, kp)

	kg.On("GiveUp").Return()
	kp.On("Output", kg, nil).Return("giveup_output")

	result := ki.GiveUp()
	assert.Equal(t, "giveup_output", result)
	kg.AssertCalled(t, "GiveUp")
}

func TestWhiteheadInteractorHint(t *testing.T) {
	kg := newMockWhiteheadGame()
	kp := newMockWhiteheadPresenter()
	ki := NewWhiteheadInteractor(kg, kp)

	kp.On("HintOutput", mock.Anything).Return("hint_output")

	result := ki.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestWhiteheadInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		kg.On("AutoComplete").Return(nil)
		kp.On("Output", kg, nil).Return("ac_output")

		result := ki.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		err := errors.New("not all face up")
		kg.On("AutoComplete").Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestWhiteheadInteractorActionLog(t *testing.T) {
	kg := newMockWhiteheadGame()
	kp := newMockWhiteheadPresenter()
	ki := NewWhiteheadInteractor(kg, kp)

	kp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := ki.ActionLog()
	assert.Equal(t, "log_output", result)
}

func TestWhiteheadInteractorResetWithConfig(t *testing.T) {
	kg := newMockWhiteheadGame()
	kp := newMockWhiteheadPresenter()
	ki := NewWhiteheadInteractor(kg, kp)

	cfg := domain.WhiteheadConfig{DrawCount: 3}
	kg.On("ResetWithConfig", cfg).Return()
	kp.On("Output", kg, nil).Return("reset_config_output")

	result := ki.ResetWithConfig(cfg)
	assert.Equal(t, "reset_config_output", result)
	kg.AssertCalled(t, "ResetWithConfig", cfg)
}

func TestWhiteheadInteractorUndo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		kg.On("Undo").Return(nil)
		kp.On("Output", kg, nil).Return("undo_output")

		result := ki.Undo()
		assert.Equal(t, "undo_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockWhiteheadGame()
		kp := newMockWhiteheadPresenter()
		ki := NewWhiteheadInteractor(kg, kp)

		err := errors.New("nothing to undo")
		kg.On("Undo").Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.Undo()
		assert.Equal(t, "error_output", result)
	})
}

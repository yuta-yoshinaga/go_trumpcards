package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newMockKlondikeGame() *interfaces.MockKlondikeGame {
	return new(interfaces.MockKlondikeGame)
}

func newMockKlondikePresenter() *presenter.MockKlondikePresenter {
	return new(presenter.MockKlondikePresenter)
}

func TestNewKlondikeInteractor(t *testing.T) {
	kg := newMockKlondikeGame()
	kp := newMockKlondikePresenter()
	ki := NewKlondikeInteractor(kg, kp)
	assert.NotNil(t, ki)
}

func TestNewKlondikeInteractorPanicsOnNil(t *testing.T) {
	kp := newMockKlondikePresenter()
	assert.Panics(t, func() { NewKlondikeInteractor(nil, kp) })
	kg := newMockKlondikeGame()
	assert.Panics(t, func() { NewKlondikeInteractor(kg, nil) })
}

func TestKlondikeInteractorReset(t *testing.T) {
	kg := newMockKlondikeGame()
	kp := newMockKlondikePresenter()
	ki := NewKlondikeInteractor(kg, kp)

	kg.On("Reset").Return()
	kp.On("Output", kg, nil).Return("reset_output")

	result := ki.Reset()
	assert.Equal(t, "reset_output", result)
	kg.AssertCalled(t, "Reset")
}

func TestKlondikeInteractorDraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		kg.On("Draw").Return(nil)
		kp.On("Output", kg, nil).Return("draw_output")

		result := ki.Draw()
		assert.Equal(t, "draw_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		err := errors.New("no cards")
		kg.On("Draw").Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.Draw()
		assert.Equal(t, "error_output", result)
	})
}

func TestKlondikeInteractorMoveWasteToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		kg.On("MoveWasteToTableau", 3).Return(nil)
		kp.On("Output", kg, nil).Return("move_output")

		result := ki.MoveWasteToTableau(3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		err := errors.New("cannot place")
		kg.On("MoveWasteToTableau", 3).Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.MoveWasteToTableau(3)
		assert.Equal(t, "error_output", result)
	})
}

func TestKlondikeInteractorMoveWasteToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		kg.On("MoveWasteToFoundation").Return(nil)
		kp.On("Output", kg, nil).Return("move_output")

		result := ki.MoveWasteToFoundation()
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		err := errors.New("cannot place")
		kg.On("MoveWasteToFoundation").Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.MoveWasteToFoundation()
		assert.Equal(t, "error_output", result)
	})
}

func TestKlondikeInteractorMoveTableauToTableau(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		kg.On("MoveTableauToTableau", 0, 2, 3).Return(nil)
		kp.On("Output", kg, nil).Return("move_output")

		result := ki.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		err := errors.New("invalid")
		kg.On("MoveTableauToTableau", 0, 2, 3).Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.MoveTableauToTableau(0, 2, 3)
		assert.Equal(t, "error_output", result)
	})
}

func TestKlondikeInteractorMoveTableauToFoundation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		kg.On("MoveTableauToFoundation", 2).Return(nil)
		kp.On("Output", kg, nil).Return("move_output")

		result := ki.MoveTableauToFoundation(2)
		assert.Equal(t, "move_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		err := errors.New("invalid")
		kg.On("MoveTableauToFoundation", 2).Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.MoveTableauToFoundation(2)
		assert.Equal(t, "error_output", result)
	})
}

func TestKlondikeInteractorGiveUp(t *testing.T) {
	kg := newMockKlondikeGame()
	kp := newMockKlondikePresenter()
	ki := NewKlondikeInteractor(kg, kp)

	kg.On("GiveUp").Return()
	kp.On("Output", kg, nil).Return("giveup_output")

	result := ki.GiveUp()
	assert.Equal(t, "giveup_output", result)
	kg.AssertCalled(t, "GiveUp")
}

func TestKlondikeInteractorHint(t *testing.T) {
	kg := newMockKlondikeGame()
	kp := newMockKlondikePresenter()
	ki := NewKlondikeInteractor(kg, kp)

	kp.On("HintOutput", mock.Anything).Return("hint_output")

	result := ki.Hint()
	assert.Equal(t, "hint_output", result)
}

func TestKlondikeInteractorAutoComplete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		kg.On("AutoComplete").Return(nil)
		kp.On("Output", kg, nil).Return("ac_output")

		result := ki.AutoComplete()
		assert.Equal(t, "ac_output", result)
	})

	t.Run("error", func(t *testing.T) {
		kg := newMockKlondikeGame()
		kp := newMockKlondikePresenter()
		ki := NewKlondikeInteractor(kg, kp)

		err := errors.New("not all face up")
		kg.On("AutoComplete").Return(err)
		kp.On("Output", kg, err).Return("error_output")

		result := ki.AutoComplete()
		assert.Equal(t, "error_output", result)
	})
}

func TestKlondikeInteractorActionLog(t *testing.T) {
	kg := newMockKlondikeGame()
	kp := newMockKlondikePresenter()
	ki := NewKlondikeInteractor(kg, kp)

	kp.On("ActionLogOutput", mock.Anything).Return("log_output")

	result := ki.ActionLog()
	assert.Equal(t, "log_output", result)
}

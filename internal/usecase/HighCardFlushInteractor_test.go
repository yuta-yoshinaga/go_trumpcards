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

func TestNewHighCardFlushInteractor(t *testing.T) {
	mockGame := new(interfaces.MockHighCardFlushGame)
	mockPresenter := new(presenter.MockHighCardFlushPresenter)
	hi := NewHighCardFlushInteractor(mockGame, mockPresenter)
	assert.NotNil(t, hi)
}

func TestNewHighCardFlushInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockHighCardFlushPresenter)
	assert.Panics(t, func() { NewHighCardFlushInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockHighCardFlushGame)
	assert.Panics(t, func() { NewHighCardFlushInteractor(mockGame, nil) })
}

func TestHighCardFlushInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockHighCardFlushGame)
	mockPresenter := new(presenter.MockHighCardFlushPresenter)
	hi := NewHighCardFlushInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := hi.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestHighCardFlushInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockHighCardFlushGame)
		mockPresenter := new(presenter.MockHighCardFlushPresenter)
		hi := NewHighCardFlushInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 50, 20).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := hi.Bet(100, 50, 20)
		assert.Equal(t, "bet output", result)
	})
	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockHighCardFlushGame)
		mockPresenter := new(presenter.MockHighCardFlushPresenter)
		hi := NewHighCardFlushInteractor(mockGame, mockPresenter)

		err := errors.New("bet err")
		mockGame.On("Bet", 100, 0, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("err output")

		assert.Equal(t, "err output", hi.Bet(100, 0, 0))
	})
}

func TestHighCardFlushInteractor_Raise(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockHighCardFlushGame)
		mockPresenter := new(presenter.MockHighCardFlushPresenter)
		hi := NewHighCardFlushInteractor(mockGame, mockPresenter)

		mockGame.On("Raise", 2).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("raise output")

		assert.Equal(t, "raise output", hi.Raise(2))
	})
	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockHighCardFlushGame)
		mockPresenter := new(presenter.MockHighCardFlushPresenter)
		hi := NewHighCardFlushInteractor(mockGame, mockPresenter)

		mockGame.On("Raise", 1).Return(errors.New("raise err"))
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("err")
		assert.Equal(t, "err", hi.Raise(1))
	})
}

func TestHighCardFlushInteractor_Fold(t *testing.T) {
	mockGame := new(interfaces.MockHighCardFlushGame)
	mockPresenter := new(presenter.MockHighCardFlushPresenter)
	hi := NewHighCardFlushInteractor(mockGame, mockPresenter)

	mockGame.On("Fold").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("fold output")
	assert.Equal(t, "fold output", hi.Fold())
}

func TestHighCardFlushInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockHighCardFlushGame)
	mockPresenter := new(presenter.MockHighCardFlushPresenter)
	hi := NewHighCardFlushInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log")
	assert.Equal(t, "log", hi.ActionLog())
}

func TestHighCardFlushInteractor_Hint(t *testing.T) {
	mockGame := new(interfaces.MockHighCardFlushGame)
	mockPresenter := new(presenter.MockHighCardFlushPresenter)
	hi := NewHighCardFlushInteractor(mockGame, mockPresenter)

	mockPresenter.On("HintOutput", mockGame).Return("hint")
	assert.Equal(t, "hint", hi.Hint())
}

func TestRestoreHighCardFlushInteractor(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockPresenter := new(presenter.MockHighCardFlushPresenter)
		// Marshal a fresh game and round-trip it.
		game := domain.NewDefaultHighCardFlush()
		data, err := game.MarshalJSON()
		assert.NoError(t, err)
		restored, err := RestoreHighCardFlushInteractor(data, mockPresenter)
		assert.NoError(t, err)
		assert.NotNil(t, restored)
	})
	t.Run("invalid json fails", func(t *testing.T) {
		mockPresenter := new(presenter.MockHighCardFlushPresenter)
		_, err := RestoreHighCardFlushInteractor([]byte("not json"), mockPresenter)
		assert.Error(t, err)
	})
}

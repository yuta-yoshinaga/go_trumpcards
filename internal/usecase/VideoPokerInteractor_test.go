package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewVideoPokerInteractor(t *testing.T) {
	mockGame := new(interfaces.MockVideoPokerGame)
	mockPresenter := new(presenter.MockVideoPokerPresenter)
	vi := NewVideoPokerInteractor(mockGame, mockPresenter)
	assert.NotNil(t, vi)
}

func TestNewVideoPokerInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockVideoPokerPresenter)
	assert.Panics(t, func() { NewVideoPokerInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockVideoPokerGame)
	assert.Panics(t, func() { NewVideoPokerInteractor(mockGame, nil) })
}

func TestVideoPokerInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockVideoPokerGame)
	mockPresenter := new(presenter.MockVideoPokerPresenter)
	vi := NewVideoPokerInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := vi.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
	mockPresenter.AssertCalled(t, "Output", mockGame, nil)
}

func TestVideoPokerInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockVideoPokerGame)
		mockPresenter := new(presenter.MockVideoPokerPresenter)
		vi := NewVideoPokerInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 3).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := vi.Bet(3)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockVideoPokerGame)
		mockPresenter := new(presenter.MockVideoPokerPresenter)
		vi := NewVideoPokerInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 3).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := vi.Bet(3)
		assert.Equal(t, "error output", result)
	})
}

func TestVideoPokerInteractor_Hold(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockVideoPokerGame)
		mockPresenter := new(presenter.MockVideoPokerPresenter)
		vi := NewVideoPokerInteractor(mockGame, mockPresenter)

		indices := []int{0, 2, 4}
		mockGame.On("Hold", indices).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("hold output")

		result := vi.Hold(indices)
		assert.Equal(t, "hold output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockVideoPokerGame)
		mockPresenter := new(presenter.MockVideoPokerPresenter)
		vi := NewVideoPokerInteractor(mockGame, mockPresenter)

		indices := []int{5}
		err := errors.New("invalid card")
		mockGame.On("Hold", indices).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "invalid card"
		})).Return("error output")

		result := vi.Hold(indices)
		assert.Equal(t, "error output", result)
	})
}

func TestVideoPokerInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockVideoPokerGame)
	mockPresenter := new(presenter.MockVideoPokerPresenter)
	vi := NewVideoPokerInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	result := vi.ActionLog()
	assert.Equal(t, "log output", result)
}

func TestVideoPokerInteractor_Hint(t *testing.T) {
	mockGame := new(interfaces.MockVideoPokerGame)
	mockPresenter := new(presenter.MockVideoPokerPresenter)
	vi := NewVideoPokerInteractor(mockGame, mockPresenter)

	mockPresenter.On("HintOutput", mockGame).Return("hint output")
	assert.Equal(t, "hint output", vi.Hint())
}

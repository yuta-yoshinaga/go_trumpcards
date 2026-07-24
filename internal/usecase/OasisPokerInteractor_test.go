package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewOasisPokerInteractor(t *testing.T) {
	mockGame := new(interfaces.MockOasisPokerGame)
	mockPresenter := new(presenter.MockOasisPokerPresenter)
	oi := NewOasisPokerInteractor(mockGame, mockPresenter)
	assert.NotNil(t, oi)
}

func TestNewOasisPokerInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockOasisPokerPresenter)
	assert.Panics(t, func() { NewOasisPokerInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockOasisPokerGame)
	assert.Panics(t, func() { NewOasisPokerInteractor(mockGame, nil) })
}

func TestOasisPokerInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockOasisPokerGame)
	mockPresenter := new(presenter.MockOasisPokerPresenter)
	oi := NewOasisPokerInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	assert.Equal(t, "reset output", oi.Reset())
	mockGame.AssertCalled(t, "Reset")
}

func TestOasisPokerInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockOasisPokerGame)
		mockPresenter := new(presenter.MockOasisPokerPresenter)
		oi := NewOasisPokerInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 10).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		assert.Equal(t, "bet output", oi.Bet(100, 10))
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockOasisPokerGame)
		mockPresenter := new(presenter.MockOasisPokerPresenter)
		oi := NewOasisPokerInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		assert.Equal(t, "error output", oi.Bet(100, 0))
	})
}

func TestOasisPokerInteractor_Exchange(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockOasisPokerGame)
		mockPresenter := new(presenter.MockOasisPokerPresenter)
		oi := NewOasisPokerInteractor(mockGame, mockPresenter)

		mockGame.On("Exchange", []int{0, 2}).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("exchange output")

		assert.Equal(t, "exchange output", oi.Exchange([]int{0, 2}))
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockOasisPokerGame)
		mockPresenter := new(presenter.MockOasisPokerPresenter)
		oi := NewOasisPokerInteractor(mockGame, mockPresenter)

		err := errors.New("ex error")
		mockGame.On("Exchange", mock.Anything).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil
		})).Return("ex error output")

		assert.Equal(t, "ex error output", oi.Exchange([]int{0}))
	})
}

func TestOasisPokerInteractor_Stand(t *testing.T) {
	mockGame := new(interfaces.MockOasisPokerGame)
	mockPresenter := new(presenter.MockOasisPokerPresenter)
	oi := NewOasisPokerInteractor(mockGame, mockPresenter)

	mockGame.On("Stand").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("stand output")

	assert.Equal(t, "stand output", oi.Stand())
}

func TestOasisPokerInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockOasisPokerGame)
		mockPresenter := new(presenter.MockOasisPokerPresenter)
		oi := NewOasisPokerInteractor(mockGame, mockPresenter)

		mockGame.On("Play").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("play output")

		assert.Equal(t, "play output", oi.Play())
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockOasisPokerGame)
		mockPresenter := new(presenter.MockOasisPokerPresenter)
		oi := NewOasisPokerInteractor(mockGame, mockPresenter)

		err := errors.New("play error")
		mockGame.On("Play").Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil
		})).Return("play error output")

		assert.Equal(t, "play error output", oi.Play())
	})
}

func TestOasisPokerInteractor_Fold(t *testing.T) {
	mockGame := new(interfaces.MockOasisPokerGame)
	mockPresenter := new(presenter.MockOasisPokerPresenter)
	oi := NewOasisPokerInteractor(mockGame, mockPresenter)

	mockGame.On("Fold").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("fold output")

	assert.Equal(t, "fold output", oi.Fold())
}

func TestOasisPokerInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockOasisPokerGame)
	mockPresenter := new(presenter.MockOasisPokerPresenter)
	oi := NewOasisPokerInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	assert.Equal(t, "log output", oi.ActionLog())
}

func TestOasisPokerInteractor_Hint(t *testing.T) {
	mockGame := new(interfaces.MockOasisPokerGame)
	mockPresenter := new(presenter.MockOasisPokerPresenter)
	oi := NewOasisPokerInteractor(mockGame, mockPresenter)

	mockPresenter.On("HintOutput", mockGame).Return("hint output")
	assert.Equal(t, "hint output", oi.Hint())
}

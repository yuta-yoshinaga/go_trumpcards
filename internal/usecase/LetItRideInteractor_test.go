package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewLetItRideInteractor(t *testing.T) {
	mockGame := new(interfaces.MockLetItRideGame)
	mockPresenter := new(presenter.MockLetItRidePresenter)
	li := NewLetItRideInteractor(mockGame, mockPresenter)
	assert.NotNil(t, li)
}

func TestNewLetItRideInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockLetItRidePresenter)
	assert.Panics(t, func() { NewLetItRideInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockLetItRideGame)
	assert.Panics(t, func() { NewLetItRideInteractor(mockGame, nil) })
}

func TestLetItRideInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockLetItRideGame)
	mockPresenter := new(presenter.MockLetItRidePresenter)
	li := NewLetItRideInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := li.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestLetItRideInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockLetItRideGame)
		mockPresenter := new(presenter.MockLetItRidePresenter)
		li := NewLetItRideInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := li.Bet(100)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockLetItRideGame)
		mockPresenter := new(presenter.MockLetItRidePresenter)
		li := NewLetItRideInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := li.Bet(100)
		assert.Equal(t, "error output", result)
	})
}

func TestLetItRideInteractor_Pull(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockLetItRideGame)
		mockPresenter := new(presenter.MockLetItRidePresenter)
		li := NewLetItRideInteractor(mockGame, mockPresenter)

		mockGame.On("Pull").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("pull output")

		result := li.Pull()
		assert.Equal(t, "pull output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockLetItRideGame)
		mockPresenter := new(presenter.MockLetItRidePresenter)
		li := NewLetItRideInteractor(mockGame, mockPresenter)

		err := errors.New("pull error")
		mockGame.On("Pull").Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil
		})).Return("pull error output")

		result := li.Pull()
		assert.Equal(t, "pull error output", result)
	})
}

func TestLetItRideInteractor_LetItRide(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockLetItRideGame)
		mockPresenter := new(presenter.MockLetItRidePresenter)
		li := NewLetItRideInteractor(mockGame, mockPresenter)

		mockGame.On("LetItRideAction").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("letitride output")

		result := li.LetItRide()
		assert.Equal(t, "letitride output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockLetItRideGame)
		mockPresenter := new(presenter.MockLetItRidePresenter)
		li := NewLetItRideInteractor(mockGame, mockPresenter)

		err := errors.New("letitride error")
		mockGame.On("LetItRideAction").Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil
		})).Return("letitride error output")

		result := li.LetItRide()
		assert.Equal(t, "letitride error output", result)
	})
}

func TestLetItRideInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockLetItRideGame)
	mockPresenter := new(presenter.MockLetItRidePresenter)
	li := NewLetItRideInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	result := li.ActionLog()
	assert.Equal(t, "log output", result)
}

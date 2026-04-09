package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewPaiGowInteractor(t *testing.T) {
	mockGame := new(interfaces.MockPaiGowGame)
	mockPresenter := new(presenter.MockPaiGowPresenter)
	pi := NewPaiGowInteractor(mockGame, mockPresenter)
	assert.NotNil(t, pi)
}

func TestNewPaiGowInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockPaiGowPresenter)
	assert.Panics(t, func() { NewPaiGowInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockPaiGowGame)
	assert.Panics(t, func() { NewPaiGowInteractor(mockGame, nil) })
}

func TestPaiGowInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockPaiGowGame)
	mockPresenter := new(presenter.MockPaiGowPresenter)
	pi := NewPaiGowInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := pi.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestPaiGowInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockPaiGowGame)
		mockPresenter := new(presenter.MockPaiGowPresenter)
		pi := NewPaiGowInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := pi.Bet(100)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockPaiGowGame)
		mockPresenter := new(presenter.MockPaiGowPresenter)
		pi := NewPaiGowInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := pi.Bet(100)
		assert.Equal(t, "error output", result)
	})
}

func TestPaiGowInteractor_SetHands(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockPaiGowGame)
		mockPresenter := new(presenter.MockPaiGowPresenter)
		pi := NewPaiGowInteractor(mockGame, mockPresenter)

		mockGame.On("SetHands", 0, 1).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("set output")

		result := pi.SetHands(0, 1)
		assert.Equal(t, "set output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockPaiGowGame)
		mockPresenter := new(presenter.MockPaiGowPresenter)
		pi := NewPaiGowInteractor(mockGame, mockPresenter)

		err := errors.New("invalid")
		mockGame.On("SetHands", 0, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil
		})).Return("error output")

		result := pi.SetHands(0, 0)
		assert.Equal(t, "error output", result)
	})
}

func TestPaiGowInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockPaiGowGame)
	mockPresenter := new(presenter.MockPaiGowPresenter)
	pi := NewPaiGowInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	result := pi.ActionLog()
	assert.Equal(t, "log output", result)
}

package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewCasinoHoldemInteractor(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)
	assert.NotNil(t, ci)
}

func TestNewCasinoHoldemInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	assert.Panics(t, func() { NewCasinoHoldemInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockCasinoHoldemGame)
	assert.Panics(t, func() { NewCasinoHoldemInteractor(mockGame, nil) })
}

func TestCasinoHoldemInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := ci.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestCasinoHoldemInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockCasinoHoldemGame)
		mockPresenter := new(presenter.MockCasinoHoldemPresenter)
		ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 10).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := ci.Bet(100, 10)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockCasinoHoldemGame)
		mockPresenter := new(presenter.MockCasinoHoldemPresenter)
		ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := ci.Bet(100, 0)
		assert.Equal(t, "error output", result)
	})
}

func TestCasinoHoldemInteractor_Call(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Call").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("call output")

	assert.Equal(t, "call output", ci.Call())
}

func TestCasinoHoldemInteractor_Fold(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Fold").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("fold output")

	assert.Equal(t, "fold output", ci.Fold())
}

func TestCasinoHoldemInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockCasinoHoldemGame)
	mockPresenter := new(presenter.MockCasinoHoldemPresenter)
	ci := NewCasinoHoldemInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	assert.Equal(t, "log output", ci.ActionLog())
}

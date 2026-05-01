package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewTexasHoldemBonusInteractor(t *testing.T) {
	mockGame := new(interfaces.MockTexasHoldemBonusGame)
	mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
	ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)
	assert.NotNil(t, ti)
}

func TestNewTexasHoldemBonusInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
	assert.Panics(t, func() { NewTexasHoldemBonusInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockTexasHoldemBonusGame)
	assert.Panics(t, func() { NewTexasHoldemBonusInteractor(mockGame, nil) })
}

func TestTexasHoldemBonusInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockTexasHoldemBonusGame)
	mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
	ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := ti.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestTexasHoldemBonusInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockTexasHoldemBonusGame)
		mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
		ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 10).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := ti.Bet(100, 10)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockTexasHoldemBonusGame)
		mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
		ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := ti.Bet(100, 0)
		assert.Equal(t, "error output", result)
	})
}

func TestTexasHoldemBonusInteractor_Play(t *testing.T) {
	mockGame := new(interfaces.MockTexasHoldemBonusGame)
	mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
	ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)

	mockGame.On("Play").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("play output")

	result := ti.Play()
	assert.Equal(t, "play output", result)
}

func TestTexasHoldemBonusInteractor_Fold(t *testing.T) {
	mockGame := new(interfaces.MockTexasHoldemBonusGame)
	mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
	ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)

	mockGame.On("Fold").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("fold output")

	result := ti.Fold()
	assert.Equal(t, "fold output", result)
}

func TestTexasHoldemBonusInteractor_Check(t *testing.T) {
	mockGame := new(interfaces.MockTexasHoldemBonusGame)
	mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
	ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)

	mockGame.On("Check").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("check output")

	assert.Equal(t, "check output", ti.Check())
}

func TestTexasHoldemBonusInteractor_Raise(t *testing.T) {
	mockGame := new(interfaces.MockTexasHoldemBonusGame)
	mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
	ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)

	mockGame.On("Raise").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("raise output")

	assert.Equal(t, "raise output", ti.Raise())
}

func TestTexasHoldemBonusInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockTexasHoldemBonusGame)
	mockPresenter := new(presenter.MockTexasHoldemBonusPresenter)
	ti := NewTexasHoldemBonusInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	assert.Equal(t, "log output", ti.ActionLog())
}

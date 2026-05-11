package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewUltimateTexasHoldemInteractor(t *testing.T) {
	mockGame := new(interfaces.MockUltimateTexasHoldemGame)
	mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
	ui := NewUltimateTexasHoldemInteractor(mockGame, mockPresenter)
	assert.NotNil(t, ui)
}

func TestNewUltimateTexasHoldemInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
	assert.Panics(t, func() { NewUltimateTexasHoldemInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockUltimateTexasHoldemGame)
	assert.Panics(t, func() { NewUltimateTexasHoldemInteractor(mockGame, nil) })
}

func TestUltimateTexasHoldemInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockUltimateTexasHoldemGame)
	mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
	ui := NewUltimateTexasHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	assert.Equal(t, "reset output", ui.Reset())
	mockGame.AssertCalled(t, "Reset")
}

func TestUltimateTexasHoldemInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockUltimateTexasHoldemGame)
		mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
		ui := NewUltimateTexasHoldemInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 20).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		assert.Equal(t, "bet output", ui.Bet(100, 20))
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockUltimateTexasHoldemGame)
		mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
		ui := NewUltimateTexasHoldemInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		assert.Equal(t, "error output", ui.Bet(100, 0))
	})
}

func TestUltimateTexasHoldemInteractor_Play(t *testing.T) {
	mockGame := new(interfaces.MockUltimateTexasHoldemGame)
	mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
	ui := NewUltimateTexasHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Play", 4).Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("play output")

	assert.Equal(t, "play output", ui.Play(4))
}

func TestUltimateTexasHoldemInteractor_Check(t *testing.T) {
	mockGame := new(interfaces.MockUltimateTexasHoldemGame)
	mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
	ui := NewUltimateTexasHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Check").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("check output")

	assert.Equal(t, "check output", ui.Check())
}

func TestUltimateTexasHoldemInteractor_Fold(t *testing.T) {
	mockGame := new(interfaces.MockUltimateTexasHoldemGame)
	mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
	ui := NewUltimateTexasHoldemInteractor(mockGame, mockPresenter)

	mockGame.On("Fold").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("fold output")

	assert.Equal(t, "fold output", ui.Fold())
}

func TestUltimateTexasHoldemInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockUltimateTexasHoldemGame)
	mockPresenter := new(presenter.MockUltimateTexasHoldemPresenter)
	ui := NewUltimateTexasHoldemInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("action log output")

	assert.Equal(t, "action log output", ui.ActionLog())
}

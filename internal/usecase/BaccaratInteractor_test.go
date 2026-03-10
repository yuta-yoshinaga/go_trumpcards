package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewBaccaratInteractor(t *testing.T) {
	mockGame := new(interfaces.MockBaccaratGame)
	mockPresenter := new(presenter.MockBaccaratPresenter)
	bi := NewBaccaratInteractor(mockGame, mockPresenter)
	assert.NotNil(t, bi)
}

func TestNewBaccaratInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockBaccaratPresenter)
	assert.Panics(t, func() { NewBaccaratInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockBaccaratGame)
	assert.Panics(t, func() { NewBaccaratInteractor(mockGame, nil) })
}

func TestBaccaratInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockBaccaratGame)
	mockPresenter := new(presenter.MockBaccaratPresenter)
	bi := NewBaccaratInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := bi.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
	mockPresenter.AssertCalled(t, "Output", mockGame, nil)
}

func TestBaccaratInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockBaccaratGame)
		mockPresenter := new(presenter.MockBaccaratPresenter)
		bi := NewBaccaratInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 0).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := bi.Bet(100, 0)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockBaccaratGame)
		mockPresenter := new(presenter.MockBaccaratPresenter)
		bi := NewBaccaratInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := bi.Bet(100, 0)
		assert.Equal(t, "error output", result)
	})
}

func TestBaccaratInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockBaccaratGame)
	mockPresenter := new(presenter.MockBaccaratPresenter)
	bi := NewBaccaratInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	result := bi.ActionLog()
	assert.Equal(t, "log output", result)
}

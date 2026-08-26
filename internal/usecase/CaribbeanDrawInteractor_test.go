package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewCaribbeanDrawInteractor(t *testing.T) {
	mockGame := new(interfaces.MockCaribbeanDrawGame)
	mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
	ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)
	assert.NotNil(t, ci)
}

func TestNewCaribbeanDrawInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
	assert.Panics(t, func() { NewCaribbeanDrawInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockCaribbeanDrawGame)
	assert.Panics(t, func() { NewCaribbeanDrawInteractor(mockGame, nil) })
}

func TestCaribbeanDrawInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockCaribbeanDrawGame)
	mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
	ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := ci.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestCaribbeanDrawInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockCaribbeanDrawGame)
		mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
		ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100, 10).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := ci.Bet(100, 10)
		assert.Equal(t, "bet output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockCaribbeanDrawGame)
		mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
		ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)

		err := errors.New("test error")
		mockGame.On("Bet", 100, 0).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil && e.Error() == "test error"
		})).Return("error output")

		result := ci.Bet(100, 0)
		assert.Equal(t, "error output", result)
	})
}

func TestCaribbeanDrawInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockCaribbeanDrawGame)
		mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
		ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)

		mockGame.On("Play").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("play output")

		result := ci.Play()
		assert.Equal(t, "play output", result)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockCaribbeanDrawGame)
		mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
		ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)

		err := errors.New("play error")
		mockGame.On("Play").Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool {
			return e != nil
		})).Return("play error output")

		result := ci.Play()
		assert.Equal(t, "play error output", result)
	})
}

func TestCaribbeanDrawInteractor_Fold(t *testing.T) {
	mockGame := new(interfaces.MockCaribbeanDrawGame)
	mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
	ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)

	mockGame.On("Fold").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("fold output")

	result := ci.Fold()
	assert.Equal(t, "fold output", result)
}

func TestCaribbeanDrawInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockCaribbeanDrawGame)
	mockPresenter := new(presenter.MockCaribbeanDrawPresenter)
	ci := NewCaribbeanDrawInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")

	result := ci.ActionLog()
	assert.Equal(t, "log output", result)
}

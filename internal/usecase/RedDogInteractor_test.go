package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewRedDogInteractor(t *testing.T) {
	mockGame := new(interfaces.MockRedDogGame)
	mockPresenter := new(presenter.MockRedDogPresenter)
	ri := NewRedDogInteractor(mockGame, mockPresenter)
	assert.NotNil(t, ri)
}

func TestNewRedDogInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockRedDogPresenter)
	assert.Panics(t, func() { NewRedDogInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockRedDogGame)
	assert.Panics(t, func() { NewRedDogInteractor(mockGame, nil) })
}

func TestRedDogInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockRedDogGame)
	mockPresenter := new(presenter.MockRedDogPresenter)
	ri := NewRedDogInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	result := ri.Reset()
	assert.Equal(t, "reset output", result)
	mockGame.AssertCalled(t, "Reset")
}

func TestRedDogInteractor_Bet(t *testing.T) {
	t.Run("success calls ResolveInitial", func(t *testing.T) {
		mockGame := new(interfaces.MockRedDogGame)
		mockPresenter := new(presenter.MockRedDogPresenter)
		ri := NewRedDogInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100).Return(nil)
		mockGame.On("ResolveInitial").Return()
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		result := ri.Bet(100)
		assert.Equal(t, "bet output", result)
		mockGame.AssertCalled(t, "ResolveInitial")
	})

	t.Run("error skips ResolveInitial", func(t *testing.T) {
		mockGame := new(interfaces.MockRedDogGame)
		mockPresenter := new(presenter.MockRedDogPresenter)
		ri := NewRedDogInteractor(mockGame, mockPresenter)

		err := errors.New("bet error")
		mockGame.On("Bet", 100).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("error output")

		result := ri.Bet(100)
		assert.Equal(t, "error output", result)
		mockGame.AssertNotCalled(t, "ResolveInitial")
	})
}

func TestRedDogInteractor_Raise(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockRedDogGame)
		mockPresenter := new(presenter.MockRedDogPresenter)
		ri := NewRedDogInteractor(mockGame, mockPresenter)

		mockGame.On("Raise", 50).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("raise output")

		assert.Equal(t, "raise output", ri.Raise(50))
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockRedDogGame)
		mockPresenter := new(presenter.MockRedDogPresenter)
		ri := NewRedDogInteractor(mockGame, mockPresenter)

		mockGame.On("Raise", 50).Return(errors.New("raise error"))
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("raise err")

		assert.Equal(t, "raise err", ri.Raise(50))
	})
}

func TestRedDogInteractor_Stay(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockRedDogGame)
		mockPresenter := new(presenter.MockRedDogPresenter)
		ri := NewRedDogInteractor(mockGame, mockPresenter)

		mockGame.On("Stay").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("stay output")

		assert.Equal(t, "stay output", ri.Stay())
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockRedDogGame)
		mockPresenter := new(presenter.MockRedDogPresenter)
		ri := NewRedDogInteractor(mockGame, mockPresenter)

		mockGame.On("Stay").Return(errors.New("stay error"))
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("stay err")

		assert.Equal(t, "stay err", ri.Stay())
	})
}

func TestRedDogInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockRedDogGame)
	mockPresenter := new(presenter.MockRedDogPresenter)
	ri := NewRedDogInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")
	assert.Equal(t, "log output", ri.ActionLog())
}

func TestRestoreRedDogInteractor(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		_, err := RestoreRedDogInteractor([]byte("not json"), nil)
		assert.Error(t, err)
	})
}

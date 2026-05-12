package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewMississippiStudInteractor(t *testing.T) {
	mockGame := new(interfaces.MockMississippiStudGame)
	mockPresenter := new(presenter.MockMississippiStudPresenter)
	mi := NewMississippiStudInteractor(mockGame, mockPresenter)
	assert.NotNil(t, mi)
}

func TestNewMississippiStudInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockMississippiStudPresenter)
	assert.Panics(t, func() { NewMississippiStudInteractor(nil, mockPresenter) })

	mockGame := new(interfaces.MockMississippiStudGame)
	assert.Panics(t, func() { NewMississippiStudInteractor(mockGame, nil) })
}

func TestMississippiStudInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockMississippiStudGame)
	mockPresenter := new(presenter.MockMississippiStudPresenter)
	mi := NewMississippiStudInteractor(mockGame, mockPresenter)

	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")

	got := mi.Reset()
	assert.Equal(t, "reset output", got)
	mockGame.AssertCalled(t, "Reset")
}

func TestMississippiStudInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockMississippiStudGame)
		mockPresenter := new(presenter.MockMississippiStudPresenter)
		mi := NewMississippiStudInteractor(mockGame, mockPresenter)

		mockGame.On("Bet", 100).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")

		got := mi.Bet(100)
		assert.Equal(t, "bet output", got)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockMississippiStudGame)
		mockPresenter := new(presenter.MockMississippiStudPresenter)
		mi := NewMississippiStudInteractor(mockGame, mockPresenter)

		err := errors.New("bet error")
		mockGame.On("Bet", 100).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("err output")

		got := mi.Bet(100)
		assert.Equal(t, "err output", got)
	})
}

func TestMississippiStudInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockMississippiStudGame)
		mockPresenter := new(presenter.MockMississippiStudPresenter)
		mi := NewMississippiStudInteractor(mockGame, mockPresenter)

		mockGame.On("Play", 3).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("play output")

		got := mi.Play(3)
		assert.Equal(t, "play output", got)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockMississippiStudGame)
		mockPresenter := new(presenter.MockMississippiStudPresenter)
		mi := NewMississippiStudInteractor(mockGame, mockPresenter)

		err := errors.New("play error")
		mockGame.On("Play", 2).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("err output")

		got := mi.Play(2)
		assert.Equal(t, "err output", got)
	})
}

func TestMississippiStudInteractor_Fold(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockMississippiStudGame)
		mockPresenter := new(presenter.MockMississippiStudPresenter)
		mi := NewMississippiStudInteractor(mockGame, mockPresenter)

		mockGame.On("Fold").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("fold output")

		got := mi.Fold()
		assert.Equal(t, "fold output", got)
	})

	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockMississippiStudGame)
		mockPresenter := new(presenter.MockMississippiStudPresenter)
		mi := NewMississippiStudInteractor(mockGame, mockPresenter)

		err := errors.New("fold error")
		mockGame.On("Fold").Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("err output")

		got := mi.Fold()
		assert.Equal(t, "err output", got)
	})
}

func TestMississippiStudInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockMississippiStudGame)
	mockPresenter := new(presenter.MockMississippiStudPresenter)
	mi := NewMississippiStudInteractor(mockGame, mockPresenter)

	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")
	got := mi.ActionLog()
	assert.Equal(t, "log output", got)
}

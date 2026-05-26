package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewRussianPokerInteractor(t *testing.T) {
	mockGame := new(interfaces.MockRussianPokerGame)
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	ri := NewRussianPokerInteractor(mockGame, mockPresenter)
	assert.NotNil(t, ri)
}

func TestNewRussianPokerInteractor_NilPanics(t *testing.T) {
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	assert.Panics(t, func() { NewRussianPokerInteractor(nil, mockPresenter) })
	mockGame := new(interfaces.MockRussianPokerGame)
	assert.Panics(t, func() { NewRussianPokerInteractor(mockGame, nil) })
}

func TestRussianPokerInteractor_Reset(t *testing.T) {
	mockGame := new(interfaces.MockRussianPokerGame)
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	ri := NewRussianPokerInteractor(mockGame, mockPresenter)
	mockGame.On("Reset").Return()
	mockPresenter.On("Output", mockGame, nil).Return("reset output")
	assert.Equal(t, "reset output", ri.Reset())
	mockGame.AssertCalled(t, "Reset")
}

func TestRussianPokerInteractor_Bet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockRussianPokerGame)
		mockPresenter := new(presenter.MockRussianPokerPresenter)
		ri := NewRussianPokerInteractor(mockGame, mockPresenter)
		mockGame.On("Bet", 100).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("bet output")
		assert.Equal(t, "bet output", ri.Bet(100))
	})
	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockRussianPokerGame)
		mockPresenter := new(presenter.MockRussianPokerPresenter)
		ri := NewRussianPokerInteractor(mockGame, mockPresenter)
		err := errors.New("test error")
		mockGame.On("Bet", 100).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("error output")
		assert.Equal(t, "error output", ri.Bet(100))
	})
}

func TestRussianPokerInteractor_Exchange(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockRussianPokerGame)
		mockPresenter := new(presenter.MockRussianPokerPresenter)
		ri := NewRussianPokerInteractor(mockGame, mockPresenter)
		mockGame.On("Exchange", []int{0, 2}).Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("exchange output")
		assert.Equal(t, "exchange output", ri.Exchange([]int{0, 2}))
	})
	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockRussianPokerGame)
		mockPresenter := new(presenter.MockRussianPokerPresenter)
		ri := NewRussianPokerInteractor(mockGame, mockPresenter)
		err := errors.New("ex error")
		mockGame.On("Exchange", mock.Anything).Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("ex error output")
		assert.Equal(t, "ex error output", ri.Exchange([]int{0}))
	})
}

func TestRussianPokerInteractor_Buy6th(t *testing.T) {
	mockGame := new(interfaces.MockRussianPokerGame)
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	ri := NewRussianPokerInteractor(mockGame, mockPresenter)
	mockGame.On("Buy6th").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("buy6th output")
	assert.Equal(t, "buy6th output", ri.Buy6th())
}

func TestRussianPokerInteractor_Select(t *testing.T) {
	mockGame := new(interfaces.MockRussianPokerGame)
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	ri := NewRussianPokerInteractor(mockGame, mockPresenter)
	mockGame.On("Select", 3).Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("select output")
	assert.Equal(t, "select output", ri.Select(3))
}

func TestRussianPokerInteractor_Play(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockGame := new(interfaces.MockRussianPokerGame)
		mockPresenter := new(presenter.MockRussianPokerPresenter)
		ri := NewRussianPokerInteractor(mockGame, mockPresenter)
		mockGame.On("Play").Return(nil)
		mockPresenter.On("Output", mockGame, nil).Return("play output")
		assert.Equal(t, "play output", ri.Play())
	})
	t.Run("error", func(t *testing.T) {
		mockGame := new(interfaces.MockRussianPokerGame)
		mockPresenter := new(presenter.MockRussianPokerPresenter)
		ri := NewRussianPokerInteractor(mockGame, mockPresenter)
		err := errors.New("play error")
		mockGame.On("Play").Return(err)
		mockPresenter.On("Output", mockGame, mock.MatchedBy(func(e error) bool { return e != nil })).Return("play error output")
		assert.Equal(t, "play error output", ri.Play())
	})
}

func TestRussianPokerInteractor_Fold(t *testing.T) {
	mockGame := new(interfaces.MockRussianPokerGame)
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	ri := NewRussianPokerInteractor(mockGame, mockPresenter)
	mockGame.On("Fold").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("fold output")
	assert.Equal(t, "fold output", ri.Fold())
}

func TestRussianPokerInteractor_ForceExchange(t *testing.T) {
	mockGame := new(interfaces.MockRussianPokerGame)
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	ri := NewRussianPokerInteractor(mockGame, mockPresenter)
	mockGame.On("ForceExchange").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("force output")
	assert.Equal(t, "force output", ri.ForceExchange())
}

func TestRussianPokerInteractor_Decline(t *testing.T) {
	mockGame := new(interfaces.MockRussianPokerGame)
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	ri := NewRussianPokerInteractor(mockGame, mockPresenter)
	mockGame.On("Decline").Return(nil)
	mockPresenter.On("Output", mockGame, nil).Return("decline output")
	assert.Equal(t, "decline output", ri.Decline())
}

func TestRussianPokerInteractor_ActionLog(t *testing.T) {
	mockGame := new(interfaces.MockRussianPokerGame)
	mockPresenter := new(presenter.MockRussianPokerPresenter)
	ri := NewRussianPokerInteractor(mockGame, mockPresenter)
	mockPresenter.On("ActionLogOutput", mockGame).Return("log output")
	assert.Equal(t, "log output", ri.ActionLog())
}

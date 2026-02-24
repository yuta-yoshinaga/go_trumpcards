package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newTestPoker() *domain.Poker {
	return domain.NewPoker(domain.NewTrumpCards(0), domain.NewPokerPlayer(), domain.NewPokerPlayer())
}

func TestNewPokerInteractor_NilGuards(t *testing.T) {
	ppMock := new(presenter.MockPokerPresenter)
	t.Run("panics when p is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PokerInteractor: p must not be nil", func() {
			usecase.NewPokerInteractor(nil, ppMock)
		})
	})
	t.Run("panics when pp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PokerInteractor: pp must not be nil", func() {
			usecase.NewPokerInteractor(newTestPoker(), nil)
		})
	})
}

func TestPokerInteractor_Method(t *testing.T) {
	mockOutput := `{"dealer":{"handRank":0,"handName":"High Card","cards":[]},"player":{"handRank":0,"handName":"High Card","cards":[]},"phase":1,"message":"","pot":20,"ante":10}`
	ppMock := new(presenter.MockPokerPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	tpi := usecase.NewPokerInteractor(newTestPoker(), ppMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Reset())
	})
	t.Run("success Exchange", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Exchange([]int{0, 1}))
	})
	t.Run("success Exchange empty indices", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Exchange([]int{}))
	})
	t.Run("success Stand", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Stand())
	})
	t.Run("success Bet", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Bet(10))
	})
	t.Run("success Call", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Call())
	})
	t.Run("success Raise", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Raise(20))
	})
	t.Run("success Fold", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Fold())
	})
	t.Run("success Check", func(t *testing.T) {
		assert.Equal(t, mockOutput, tpi.Check())
	})
}

func TestPokerInteractor_MockGame(t *testing.T) {
	mockOutput := `{"phase":0}`
	ppMock := new(presenter.MockPokerPresenter)
	ppMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockPokerGame)
	gameMock.On("Reset").Return()
	gameMock.On("PlayerBet", mock.Anything).Return(nil)
	gameMock.On("PlayerCall").Return(nil)
	gameMock.On("PlayerRaise", mock.Anything).Return(nil)
	gameMock.On("PlayerFold").Return(nil)
	gameMock.On("PlayerCheck").Return(nil)
	gameMock.On("PlayerExchange", mock.Anything).Return(nil)
	gameMock.On("PlayerStand").Return(nil)

	pi := usecase.NewPokerInteractor(gameMock, ppMock)

	t.Run("Reset calls game.Reset", func(t *testing.T) {
		result := pi.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
	t.Run("Bet calls game.PlayerBet", func(t *testing.T) {
		result := pi.Bet(100)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerBet", 100)
	})
	t.Run("Call calls game.PlayerCall", func(t *testing.T) {
		result := pi.Call()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerCall")
	})
	t.Run("Raise calls game.PlayerRaise", func(t *testing.T) {
		result := pi.Raise(50)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerRaise", 50)
	})
	t.Run("Fold calls game.PlayerFold", func(t *testing.T) {
		result := pi.Fold()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerFold")
	})
	t.Run("Check calls game.PlayerCheck", func(t *testing.T) {
		result := pi.Check()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerCheck")
	})
	t.Run("Exchange calls game.PlayerExchange", func(t *testing.T) {
		result := pi.Exchange([]int{0, 2})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerExchange", []int{0, 2})
	})
	t.Run("Stand calls game.PlayerStand", func(t *testing.T) {
		result := pi.Stand()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerStand")
	})
}

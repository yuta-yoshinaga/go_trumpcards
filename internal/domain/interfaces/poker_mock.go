package interfaces

import (
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPokerGame ポーカーゲームモック
type MockPokerGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPokerGame) Reset() {
	_m.Called()
}

// PlayerExchange モック
func (_m *MockPokerGame) PlayerExchange(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// PlayerStand モック
func (_m *MockPokerGame) PlayerStand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerBet モック
func (_m *MockPokerGame) PlayerBet(amount int) error {
	ret := _m.Called(amount)
	return ret.Error(0)
}

// PlayerCall モック
func (_m *MockPokerGame) PlayerCall() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerRaise モック
func (_m *MockPokerGame) PlayerRaise(amount int) error {
	ret := _m.Called(amount)
	return ret.Error(0)
}

// PlayerFold モック
func (_m *MockPokerGame) PlayerFold() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerCheck モック
func (_m *MockPokerGame) PlayerCheck() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetPlayer モック
func (_m *MockPokerGame) GetPlayer() *domain.PokerPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.PokerPlayer); ok {
		return val
	}
	return nil
}

// GetDealer モック
func (_m *MockPokerGame) GetDealer() *domain.PokerPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.PokerPlayer); ok {
		return val
	}
	return nil
}

// GetPot モック
func (_m *MockPokerGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayerBet モック
func (_m *MockPokerGame) GetPlayerBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDealerBet モック
func (_m *MockPokerGame) GetDealerBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPhase モック
func (_m *MockPokerGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetFolded モック
func (_m *MockPokerGame) GetFolded() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetAnte モック
func (_m *MockPokerGame) GetAnte() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GameJudgment モック
func (_m *MockPokerGame) GameJudgment() int {
	ret := _m.Called()
	return ret.Int(0)
}

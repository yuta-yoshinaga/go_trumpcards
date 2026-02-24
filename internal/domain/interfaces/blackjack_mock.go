package interfaces

import (
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBlackJackGame ブラックジャックゲームモック
type MockBlackJackGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockBlackJackGame) Reset() {
	_m.Called()
}

// PlayerBet モック
func (_m *MockBlackJackGame) PlayerBet(amount int) error {
	ret := _m.Called(amount)
	return ret.Error(0)
}

// PlayerInsurance モック
func (_m *MockBlackJackGame) PlayerInsurance() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerDeclineInsurance モック
func (_m *MockBlackJackGame) PlayerDeclineInsurance() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerHit モック
func (_m *MockBlackJackGame) PlayerHit() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerStand モック
func (_m *MockBlackJackGame) PlayerStand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerDoubleDown モック
func (_m *MockBlackJackGame) PlayerDoubleDown() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerSplit モック
func (_m *MockBlackJackGame) PlayerSplit() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetPlayer モック
func (_m *MockBlackJackGame) GetPlayer() *domain.BlackJackPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BlackJackPlayer); ok {
		return val
	}
	return nil
}

// GetDealer モック
func (_m *MockBlackJackGame) GetDealer() *domain.BlackJackPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BlackJackPlayer); ok {
		return val
	}
	return nil
}

// GetPhase モック
func (_m *MockBlackJackGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag モック
func (_m *MockBlackJackGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerHands モック
func (_m *MockBlackJackGame) GetPlayerHands() []*domain.BlackJackHand {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.BlackJackHand); ok {
		return val
	}
	return nil
}

// GetCurrentHandIdx モック
func (_m *MockBlackJackGame) GetCurrentHandIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetInsuranceBet モック
func (_m *MockBlackJackGame) GetInsuranceBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsInsuranceAvailable モック
func (_m *MockBlackJackGame) IsInsuranceAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GameJudgmentForHand モック
func (_m *MockBlackJackGame) GameJudgmentForHand(handIdx int) domain.GameResult {
	ret := _m.Called(handIdx)
	return ret.Get(0).(domain.GameResult)
}

// GameJudgment モック
func (_m *MockBlackJackGame) GameJudgment() domain.GameResult {
	ret := _m.Called()
	return ret.Get(0).(domain.GameResult)
}

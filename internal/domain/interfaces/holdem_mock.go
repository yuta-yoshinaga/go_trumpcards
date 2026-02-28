package interfaces

import (
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHoldemGame テキサスホールデムゲームモック
type MockHoldemGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockHoldemGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerAction モック
func (_m *MockHoldemGame) PlayerAction(action, amount int) error {
	ret := _m.Called(action, amount)
	return ret.Error(0)
}

// GetPhase モック
func (_m *MockHoldemGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayers モック
func (_m *MockHoldemGame) GetPlayers() []*domain.HoldemPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.HoldemPlayer); ok {
		return val
	}
	return nil
}

// GetPlayer モック
func (_m *MockHoldemGame) GetPlayer(i int) *domain.HoldemPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.HoldemPlayer); ok {
		return val
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockHoldemGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCommunityCards モック
func (_m *MockHoldemGame) GetCommunityCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetPot モック
func (_m *MockHoldemGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetSidePots モック
func (_m *MockHoldemGame) GetSidePots() []domain.HoldemSidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemSidePot); ok {
		return val
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockHoldemGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockHoldemGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag モック
func (_m *MockHoldemGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastBet モック
func (_m *MockHoldemGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetMinRaise モック
func (_m *MockHoldemGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundResults モック
func (_m *MockHoldemGame) GetRoundResults() []domain.HoldemResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemResult); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockHoldemGame) GetCpuActions() []domain.HoldemCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemCpuAction); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockHoldemGame) GetConfig() domain.HoldemConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.HoldemConfig)
}

// SetConfig モック
func (_m *MockHoldemGame) SetConfig(cfg domain.HoldemConfig) {
	_m.Called(cfg)
}

// IsHumanTurn モック
func (_m *MockHoldemGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActedFlags モック
func (_m *MockHoldemGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

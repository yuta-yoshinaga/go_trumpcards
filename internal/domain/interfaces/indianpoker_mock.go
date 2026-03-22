package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockIndianPokerGame インディアンポーカーゲームモック
type MockIndianPokerGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockIndianPokerGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerAction モック
func (_m *MockIndianPokerGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// GetPhase モック
func (_m *MockIndianPokerGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayers モック
func (_m *MockIndianPokerGame) GetPlayers() []*domain.IndianPokerPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.IndianPokerPlayer); ok {
		return val
	}
	return nil
}

// GetPlayer モック
func (_m *MockIndianPokerGame) GetPlayer(i int) *domain.IndianPokerPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.IndianPokerPlayer); ok {
		return val
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockIndianPokerGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPot モック
func (_m *MockIndianPokerGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetSidePots モック
func (_m *MockIndianPokerGame) GetSidePots() []domain.IndianPokerSidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.IndianPokerSidePot); ok {
		return val
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockIndianPokerGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockIndianPokerGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag モック
func (_m *MockIndianPokerGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastBet モック
func (_m *MockIndianPokerGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetMinRaise モック
func (_m *MockIndianPokerGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRaiseCount モック
func (_m *MockIndianPokerGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundResults モック
func (_m *MockIndianPokerGame) GetRoundResults() []domain.IndianPokerResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.IndianPokerResult); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockIndianPokerGame) GetCpuActions() []domain.IndianPokerCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.IndianPokerCpuAction); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockIndianPokerGame) GetConfig() domain.IndianPokerConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.IndianPokerConfig)
}

// SetConfig モック
func (_m *MockIndianPokerGame) SetConfig(cfg domain.IndianPokerConfig) {
	_m.Called(cfg)
}

// IsHumanTurn モック
func (_m *MockIndianPokerGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActedFlags モック
func (_m *MockIndianPokerGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetHandCount モック
func (_m *MockIndianPokerGame) GetHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockIndianPokerGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

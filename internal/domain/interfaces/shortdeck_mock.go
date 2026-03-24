package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShortDeckGame ショートデックホールデムゲームモック
type MockShortDeckGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockShortDeckGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerAction モック
func (_m *MockShortDeckGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// GetPhase モック
func (_m *MockShortDeckGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayers モック
func (_m *MockShortDeckGame) GetPlayers() []*domain.ShortDeckPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ShortDeckPlayer); ok {
		return val
	}
	return nil
}

// GetPlayer モック
func (_m *MockShortDeckGame) GetPlayer(i int) *domain.ShortDeckPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.ShortDeckPlayer); ok {
		return val
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockShortDeckGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCommunityCards モック
func (_m *MockShortDeckGame) GetCommunityCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetPot モック
func (_m *MockShortDeckGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetSidePots モック
func (_m *MockShortDeckGame) GetSidePots() []domain.ShortDeckSidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.ShortDeckSidePot); ok {
		return val
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockShortDeckGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockShortDeckGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag モック
func (_m *MockShortDeckGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastBet モック
func (_m *MockShortDeckGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetMinRaise モック
func (_m *MockShortDeckGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRaiseCount モック
func (_m *MockShortDeckGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundResults モック
func (_m *MockShortDeckGame) GetRoundResults() []domain.ShortDeckResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.ShortDeckResult); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockShortDeckGame) GetCpuActions() []domain.ShortDeckCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.ShortDeckCpuAction); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockShortDeckGame) GetConfig() domain.ShortDeckConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ShortDeckConfig)
}

// SetConfig モック
func (_m *MockShortDeckGame) SetConfig(cfg domain.ShortDeckConfig) {
	_m.Called(cfg)
}

// IsHumanTurn モック
func (_m *MockShortDeckGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActedFlags モック
func (_m *MockShortDeckGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetHandCount モック
func (_m *MockShortDeckGame) GetHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Resize モック
func (_m *MockShortDeckGame) Resize(players []*domain.ShortDeckPlayer) {
	_m.Called(players)
}

// Rebuy モック
func (_m *MockShortDeckGame) Rebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SkipRebuy モック
func (_m *MockShortDeckGame) SkipRebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Addon モック
func (_m *MockShortDeckGame) Addon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SkipAddon モック
func (_m *MockShortDeckGame) SkipAddon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// IsRebuyAvailable モック
func (_m *MockShortDeckGame) IsRebuyAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsAddonAvailable モック
func (_m *MockShortDeckGame) IsAddonAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRebuyCounts モック
func (_m *MockShortDeckGame) GetRebuyCounts() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

// GetAddonUsed モック
func (_m *MockShortDeckGame) GetAddonUsed() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetRebuyPhaseType モック
func (_m *MockShortDeckGame) GetRebuyPhaseType() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Muck モック
func (_m *MockShortDeckGame) Muck() error {
	ret := _m.Called()
	return ret.Error(0)
}

// ShowHand モック
func (_m *MockShortDeckGame) ShowHand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// IsMuckAvailable モック
func (_m *MockShortDeckGame) IsMuckAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetHumanProfile モック
func (_m *MockShortDeckGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BettingHumanProfile); ok {
		return val
	}
	return nil
}

// ResetProfile モック
func (_m *MockShortDeckGame) ResetProfile() {
	_m.Called()
}

// ExportProfile モック
func (_m *MockShortDeckGame) ExportProfile() interface{} {
	ret := _m.Called()
	return ret.Get(0)
}

// ImportProfile モック
func (_m *MockShortDeckGame) ImportProfile(data []byte) error {
	ret := _m.Called(data)
	return ret.Error(0)
}

// GetActionLog モック
func (_m *MockShortDeckGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

// GetEquity モック
func (_m *MockShortDeckGame) GetEquity() *domain.HoldemEquityResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.HoldemEquityResult); ok {
		return val
	}
	return nil
}

// GetPotOdds モック
func (_m *MockShortDeckGame) GetPotOdds() float64 {
	ret := _m.Called()
	return ret.Get(0).(float64)
}

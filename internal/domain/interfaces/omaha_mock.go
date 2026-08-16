//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOmahaGame オマハホールデムゲームモック
type MockOmahaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockOmahaGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerAction モック
func (_m *MockOmahaGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// GetPhase モック
func (_m *MockOmahaGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayers モック
func (_m *MockOmahaGame) GetPlayers() []*domain.OmahaPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.OmahaPlayer); ok {
		return val
	}
	return nil
}

// GetPlayer モック
func (_m *MockOmahaGame) GetPlayer(i int) *domain.OmahaPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.OmahaPlayer); ok {
		return val
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockOmahaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCommunityCards モック
func (_m *MockOmahaGame) GetCommunityCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetPot モック
func (_m *MockOmahaGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetSidePots モック
func (_m *MockOmahaGame) GetSidePots() []domain.SidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.SidePot); ok {
		return val
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockOmahaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockOmahaGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag モック
func (_m *MockOmahaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastBet モック
func (_m *MockOmahaGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetMinRaise モック
func (_m *MockOmahaGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRaiseCount モック
func (_m *MockOmahaGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundResults モック
func (_m *MockOmahaGame) GetRoundResults() []domain.HoldemResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemResult); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockOmahaGame) GetCpuActions() []domain.HoldemCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemCpuAction); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockOmahaGame) GetConfig() domain.OmahaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.OmahaConfig)
}

// SetConfig モック
func (_m *MockOmahaGame) SetConfig(cfg domain.OmahaConfig) {
	_m.Called(cfg)
}

// IsHumanTurn モック
func (_m *MockOmahaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActedFlags モック
func (_m *MockOmahaGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetHandCount モック
func (_m *MockOmahaGame) GetHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Resize モック
func (_m *MockOmahaGame) Resize(players []*domain.OmahaPlayer) {
	_m.Called(players)
}

// Rebuy モック
func (_m *MockOmahaGame) Rebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SkipRebuy モック
func (_m *MockOmahaGame) SkipRebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Addon モック
func (_m *MockOmahaGame) Addon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SkipAddon モック
func (_m *MockOmahaGame) SkipAddon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// IsRebuyAvailable モック
func (_m *MockOmahaGame) IsRebuyAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsAddonAvailable モック
func (_m *MockOmahaGame) IsAddonAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRebuyCounts モック
func (_m *MockOmahaGame) GetRebuyCounts() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

// GetAddonUsed モック
func (_m *MockOmahaGame) GetAddonUsed() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetRebuyPhaseType モック
func (_m *MockOmahaGame) GetRebuyPhaseType() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Muck モック
func (_m *MockOmahaGame) Muck() error {
	ret := _m.Called()
	return ret.Error(0)
}

// ShowHand モック
func (_m *MockOmahaGame) ShowHand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// IsMuckAvailable モック
func (_m *MockOmahaGame) IsMuckAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetHumanProfile モック
func (_m *MockOmahaGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BettingHumanProfile); ok {
		return val
	}
	return nil
}

// ResetProfile モック
func (_m *MockOmahaGame) ResetProfile() {
	_m.Called()
}

// ExportProfile モック
func (_m *MockOmahaGame) ExportProfile() interface{} {
	ret := _m.Called()
	return ret.Get(0)
}

// ImportProfile モック
func (_m *MockOmahaGame) ImportProfile(data []byte) error {
	ret := _m.Called(data)
	return ret.Error(0)
}

// GetActionLog モック
func (_m *MockOmahaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

// GetEquity モック
func (_m *MockOmahaGame) GetEquity() *domain.HoldemEquityResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.HoldemEquityResult); ok {
		return val
	}
	return nil
}

// GetPotOdds モック
func (_m *MockOmahaGame) GetPotOdds() float64 {
	ret := _m.Called()
	return ret.Get(0).(float64)
}

// GetIsHiLo モック
func (_m *MockOmahaGame) GetIsHiLo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetBoardLowOutlook モック
func (_m *MockOmahaGame) GetBoardLowOutlook() domain.OmahaBoardLowOutlook {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.OmahaBoardLowOutlook); ok {
		return val
	}
	return domain.OmahaBoardLowOutlook{}
}

// GetHoleCardCount モック
func (_m *MockOmahaGame) GetHoleCardCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

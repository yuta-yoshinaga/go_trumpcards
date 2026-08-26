//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDramahaGame ドラマハホールデムゲームモック
type MockDramahaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockDramahaGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerAction モック
func (_m *MockDramahaGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// Draw モック
func (_m *MockDramahaGame) Draw(playerIdx int, indices []int) error {
	ret := _m.Called(playerIdx, indices)
	return ret.Error(0)
}

// GetPhase モック
func (_m *MockDramahaGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayers モック
func (_m *MockDramahaGame) GetPlayers() []*domain.DramahaPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.DramahaPlayer); ok {
		return val
	}
	return nil
}

// GetPlayer モック
func (_m *MockDramahaGame) GetPlayer(i int) *domain.DramahaPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.DramahaPlayer); ok {
		return val
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockDramahaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCommunityCards モック
func (_m *MockDramahaGame) GetCommunityCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetPot モック
func (_m *MockDramahaGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetSidePots モック
func (_m *MockDramahaGame) GetSidePots() []domain.SidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.SidePot); ok {
		return val
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockDramahaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockDramahaGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag モック
func (_m *MockDramahaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastBet モック
func (_m *MockDramahaGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetMinRaise モック
func (_m *MockDramahaGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRaiseCount モック
func (_m *MockDramahaGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundResults モック
func (_m *MockDramahaGame) GetRoundResults() []domain.HoldemResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemResult); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockDramahaGame) GetCpuActions() []domain.HoldemCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemCpuAction); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockDramahaGame) GetConfig() domain.DramahaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DramahaConfig)
}

// SetConfig モック
func (_m *MockDramahaGame) SetConfig(cfg domain.DramahaConfig) {
	_m.Called(cfg)
}

// IsHumanTurn モック
func (_m *MockDramahaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActedFlags モック
func (_m *MockDramahaGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetHandCount モック
func (_m *MockDramahaGame) GetHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Resize モック
func (_m *MockDramahaGame) Resize(players []*domain.DramahaPlayer) {
	_m.Called(players)
}

// Rebuy モック
func (_m *MockDramahaGame) Rebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SkipRebuy モック
func (_m *MockDramahaGame) SkipRebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Addon モック
func (_m *MockDramahaGame) Addon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SkipAddon モック
func (_m *MockDramahaGame) SkipAddon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// IsRebuyAvailable モック
func (_m *MockDramahaGame) IsRebuyAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsAddonAvailable モック
func (_m *MockDramahaGame) IsAddonAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRebuyCounts モック
func (_m *MockDramahaGame) GetRebuyCounts() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

// GetAddonUsed モック
func (_m *MockDramahaGame) GetAddonUsed() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetRebuyPhaseType モック
func (_m *MockDramahaGame) GetRebuyPhaseType() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Muck モック
func (_m *MockDramahaGame) Muck() error {
	ret := _m.Called()
	return ret.Error(0)
}

// ShowHand モック
func (_m *MockDramahaGame) ShowHand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// IsMuckAvailable モック
func (_m *MockDramahaGame) IsMuckAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetHumanProfile モック
func (_m *MockDramahaGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BettingHumanProfile); ok {
		return val
	}
	return nil
}

// ResetProfile モック
func (_m *MockDramahaGame) ResetProfile() {
	_m.Called()
}

// ExportProfile モック
func (_m *MockDramahaGame) ExportProfile() interface{} {
	ret := _m.Called()
	return ret.Get(0)
}

// ImportProfile モック
func (_m *MockDramahaGame) ImportProfile(data []byte) error {
	ret := _m.Called(data)
	return ret.Error(0)
}

// GetActionLog モック
func (_m *MockDramahaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

// GetEquity モック
func (_m *MockDramahaGame) GetEquity() *domain.HoldemEquityResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.HoldemEquityResult); ok {
		return val
	}
	return nil
}

// GetPotOdds モック
func (_m *MockDramahaGame) GetPotOdds() float64 {
	ret := _m.Called()
	return ret.Get(0).(float64)
}

// GetHoleCardCount モック
func (_m *MockDramahaGame) GetHoleCardCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

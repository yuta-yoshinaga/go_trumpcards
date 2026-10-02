//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPineappleGame パイナップルポーカーゲーム��ック
type MockPineappleGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPineappleGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerAction モック
func (_m *MockPineappleGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// DiscardCard モック
func (_m *MockPineappleGame) DiscardCard(cardIdx int) error {
	ret := _m.Called(cardIdx)
	return ret.Error(0)
}

// DiscardCards モック
func (_m *MockPineappleGame) DiscardCards(cardIdxs []int) error {
	ret := _m.Called(cardIdxs)
	return ret.Error(0)
}

// IsDiscardPhase モック
func (_m *MockPineappleGame) IsDiscardPhase() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モ���ク
func (_m *MockPineappleGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayers モック
func (_m *MockPineappleGame) GetPlayers() []*domain.PineapplePlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.PineapplePlayer); ok {
		return val
	}
	return nil
}

// GetPlayer モック
func (_m *MockPineappleGame) GetPlayer(i int) *domain.PineapplePlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.PineapplePlayer); ok {
		return val
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockPineappleGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCommunityCards モック
func (_m *MockPineappleGame) GetCommunityCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetPot モッ��
func (_m *MockPineappleGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetSidePots モック
func (_m *MockPineappleGame) GetSidePots() []domain.SidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.SidePot); ok {
		return val
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockPineappleGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockPineappleGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag モック
func (_m *MockPineappleGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastBet モック
func (_m *MockPineappleGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetMinRaise モック
func (_m *MockPineappleGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRaiseCount モック
func (_m *MockPineappleGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundResults モック
func (_m *MockPineappleGame) GetRoundResults() []domain.HoldemResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemResult); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockPineappleGame) GetCpuActions() []domain.HoldemCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.HoldemCpuAction); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockPineappleGame) GetConfig() domain.PineappleConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PineappleConfig)
}

// SetConfig モック
func (_m *MockPineappleGame) SetConfig(cfg domain.PineappleConfig) {
	_m.Called(cfg)
}

// IsHumanTurn モック
func (_m *MockPineappleGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActedFlags モック
func (_m *MockPineappleGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetHandCount モック
func (_m *MockPineappleGame) GetHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Resize モック
func (_m *MockPineappleGame) Resize(players []*domain.PineapplePlayer) {
	_m.Called(players)
}

// GetDiscardDone モック
func (_m *MockPineappleGame) GetDiscardDone() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetInitialDealCount モック
func (_m *MockPineappleGame) GetInitialDealCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsDiscardAfterFlopBetting モック
func (_m *MockPineappleGame) IsDiscardAfterFlopBetting() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// Rebuy モック
func (_m *MockPineappleGame) Rebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SkipRebuy モック
func (_m *MockPineappleGame) SkipRebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Addon モック
func (_m *MockPineappleGame) Addon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// SkipAddon モック
func (_m *MockPineappleGame) SkipAddon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// IsRebuyAvailable モック
func (_m *MockPineappleGame) IsRebuyAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsAddonAvailable モック
func (_m *MockPineappleGame) IsAddonAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRebuyCounts モック
func (_m *MockPineappleGame) GetRebuyCounts() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

// GetAddonUsed モック
func (_m *MockPineappleGame) GetAddonUsed() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

// GetRebuyPhaseType モック
func (_m *MockPineappleGame) GetRebuyPhaseType() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Muck モッ��
func (_m *MockPineappleGame) Muck() error {
	ret := _m.Called()
	return ret.Error(0)
}

// ShowHand モック
func (_m *MockPineappleGame) ShowHand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// IsMuckAvailable モック
func (_m *MockPineappleGame) IsMuckAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetHumanProfile モック
func (_m *MockPineappleGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BettingHumanProfile); ok {
		return val
	}
	return nil
}

// ResetProfile モック
func (_m *MockPineappleGame) ResetProfile() {
	_m.Called()
}

// ExportProfile モック
func (_m *MockPineappleGame) ExportProfile() interface{} {
	ret := _m.Called()
	return ret.Get(0)
}

// ImportProfile モック
func (_m *MockPineappleGame) ImportProfile(data []byte) error {
	ret := _m.Called(data)
	return ret.Error(0)
}

// GetActionLog モック
func (_m *MockPineappleGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

// GetHumanDiscardPairPreviews モック
func (_m *MockPineappleGame) GetHumanDiscardPairPreviews() []domain.PineappleDiscardPairPreview {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).([]domain.PineappleDiscardPairPreview)
}

// GetHumanDiscardPreviews モック
func (_m *MockPineappleGame) GetHumanDiscardPreviews() []domain.PineappleDiscardPreview {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).([]domain.PineappleDiscardPreview)
}

// GetEquity モック
func (_m *MockPineappleGame) GetEquity() *domain.HoldemEquityResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.HoldemEquityResult); ok {
		return val
	}
	return nil
}

// GetPotOdds モック
func (_m *MockPineappleGame) GetPotOdds() float64 {
	ret := _m.Called()
	return ret.Get(0).(float64)
}

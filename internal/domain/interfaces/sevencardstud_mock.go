//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSevenCardStudGame セブンカードスタッドゲームモック
type MockSevenCardStudGame struct {
	mock.Mock
}

func (_m *MockSevenCardStudGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetPlayers() []*domain.SevenCardStudPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.SevenCardStudPlayer); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetPlayer(i int) *domain.SevenCardStudPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.SevenCardStudPlayer); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetCommunityCard() *domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.Card); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetSidePots() []domain.SidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.SidePot); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSevenCardStudGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetRoundResults() []domain.SevenCardStudResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.SevenCardStudResult); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetCpuActions() []domain.SevenCardStudCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.SevenCardStudCpuAction); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetConfig() domain.SevenCardStudConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SevenCardStudConfig)
}

func (_m *MockSevenCardStudGame) SetConfig(cfg domain.SevenCardStudConfig) {
	_m.Called(cfg)
}

func (_m *MockSevenCardStudGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSevenCardStudGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) Resize(players []*domain.SevenCardStudPlayer) {
	_m.Called(players)
}

func (_m *MockSevenCardStudGame) Rebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) SkipRebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) Addon() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) SkipAddon() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) IsRebuyAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSevenCardStudGame) IsAddonAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSevenCardStudGame) GetRebuyCounts() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetAddonUsed() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetRebuyPhaseType() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) Muck() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) ShowHand() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) IsMuckAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSevenCardStudGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BettingHumanProfile); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) ResetProfile() {
	_m.Called()
}

func (_m *MockSevenCardStudGame) ExportProfile() interface{} {
	ret := _m.Called()
	return ret.Get(0)
}

func (_m *MockSevenCardStudGame) ImportProfile(data []byte) error {
	ret := _m.Called(data)
	return ret.Error(0)
}

func (_m *MockSevenCardStudGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

func (_m *MockSevenCardStudGame) GetBringInPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSevenCardStudGame) GetIsLowball() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSevenCardStudGame) GetIsHiLo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

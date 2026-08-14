//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFiveCardStudGame ファイブカードスタッドゲームモック
type MockFiveCardStudGame struct {
	mock.Mock
}

func (_m *MockFiveCardStudGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) GetIsSoko() bool { return _m.Called().Bool(0) }

func (_m *MockFiveCardStudGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) GetPlayers() []*domain.FiveCardStudPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.FiveCardStudPlayer); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetPlayer(i int) *domain.FiveCardStudPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.FiveCardStudPlayer); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) GetCommunityCard() *domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.Card); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) GetSidePots() []domain.SidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.SidePot); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFiveCardStudGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) GetRoundResults() []domain.FiveCardStudResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.FiveCardStudResult); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetCpuActions() []domain.FiveCardStudCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.FiveCardStudCpuAction); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetConfig() domain.FiveCardStudConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.FiveCardStudConfig)
}

func (_m *MockFiveCardStudGame) SetConfig(cfg domain.FiveCardStudConfig) {
	_m.Called(cfg)
}

func (_m *MockFiveCardStudGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFiveCardStudGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) Resize(players []*domain.FiveCardStudPlayer) {
	_m.Called(players)
}

func (_m *MockFiveCardStudGame) Rebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) SkipRebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) Addon() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) SkipAddon() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) IsRebuyAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFiveCardStudGame) IsAddonAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFiveCardStudGame) GetRebuyCounts() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetAddonUsed() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetRebuyPhaseType() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFiveCardStudGame) Muck() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) ShowHand() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) IsMuckAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFiveCardStudGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BettingHumanProfile); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) ResetProfile() {
	_m.Called()
}

func (_m *MockFiveCardStudGame) ExportProfile() interface{} {
	ret := _m.Called()
	return ret.Get(0)
}

func (_m *MockFiveCardStudGame) ImportProfile(data []byte) error {
	ret := _m.Called(data)
	return ret.Error(0)
}

func (_m *MockFiveCardStudGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

func (_m *MockFiveCardStudGame) GetBringInPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

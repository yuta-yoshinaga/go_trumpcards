//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFollowTheQueenGame フォロー・ザ・クイーンゲームモック
type MockFollowTheQueenGame struct {
	mock.Mock
}

func (_m *MockFollowTheQueenGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetPlayers() []*domain.FollowTheQueenPlayer {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.FollowTheQueenPlayer); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetPlayer(i int) *domain.FollowTheQueenPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.FollowTheQueenPlayer); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetCommunityCard() *domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.Card); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) IsWild(card *domain.Card) bool {
	ret := _m.Called(card)
	return ret.Bool(0)
}

func (_m *MockFollowTheQueenGame) GetWildRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetSidePots() []domain.SidePot {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.SidePot); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFollowTheQueenGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) GetRoundResults() []domain.FollowTheQueenResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.FollowTheQueenResult); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetCpuActions() []domain.FollowTheQueenCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]domain.FollowTheQueenCpuAction); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetConfig() domain.FollowTheQueenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.FollowTheQueenConfig)
}

func (_m *MockFollowTheQueenGame) SetConfig(cfg domain.FollowTheQueenConfig) {
	_m.Called(cfg)
}

func (_m *MockFollowTheQueenGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFollowTheQueenGame) GetActedFlags() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetHandCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) Resize(players []*domain.FollowTheQueenPlayer) {
	_m.Called(players)
}

func (_m *MockFollowTheQueenGame) Rebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) SkipRebuy() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) Addon() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) SkipAddon() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) IsRebuyAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFollowTheQueenGame) IsAddonAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFollowTheQueenGame) GetRebuyCounts() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetAddonUsed() []bool {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]bool); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetRebuyPhaseType() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFollowTheQueenGame) Muck() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) ShowHand() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) IsMuckAvailable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFollowTheQueenGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BettingHumanProfile); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) ResetProfile() {
	_m.Called()
}

func (_m *MockFollowTheQueenGame) ExportProfile() interface{} {
	ret := _m.Called()
	return ret.Get(0)
}

func (_m *MockFollowTheQueenGame) ImportProfile(data []byte) error {
	ret := _m.Called(data)
	return ret.Error(0)
}

func (_m *MockFollowTheQueenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

func (_m *MockFollowTheQueenGame) GetBringInPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

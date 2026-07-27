//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKingGame はキング (King) のゲームモック。
type MockKingGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockKingGame) Reset() {
	_m.Called()
}

// NextDeal モック
func (_m *MockKingGame) NextDeal() {
	_m.Called()
}

// SelectContract モック
func (_m *MockKingGame) SelectContract(contract, trumpSuit int) error {
	ret := _m.Called(contract, trumpSuit)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockKingGame) PlayerPlay(handIdx int) error {
	ret := _m.Called(handIdx)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockKingGame) CpuPlay() {
	_m.Called()
}

// GetConfig モック
func (_m *MockKingGame) GetConfig() domain.KingConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KingConfig)
}

// SetConfig モック
func (_m *MockKingGame) SetConfig(cfg domain.KingConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockKingGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockKingGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockKingGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetDealNumber モック
func (_m *MockKingGame) GetDealNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockKingGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTurn モック
func (_m *MockKingGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentContract モック
func (_m *MockKingGame) GetCurrentContract() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockKingGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockKingGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockKingGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLastTrick モック
func (_m *MockKingGame) GetLastTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLastTrickWinner モック
func (_m *MockKingGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetUsedContracts モック
func (_m *MockKingGame) GetUsedContracts() [domain.KingContractCnt]bool {
	ret := _m.Called()
	return ret.Get(0).([domain.KingContractCnt]bool)
}

// GetLastDealDetail モック
func (_m *MockKingGame) GetLastDealDetail() *domain.KingDealDetail {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.KingDealDetail)
	}
	return nil
}

// GetRoundWinners モック
func (_m *MockKingGame) GetRoundWinners() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockKingGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockKingGame) GetPlayer(i int) *domain.KingPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.KingPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockKingGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockKingGame) GetHint() *domain.KingHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.KingHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockKingGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

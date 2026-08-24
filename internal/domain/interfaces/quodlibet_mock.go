//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockQuodlibetGame はクオドリベットのゲームモック。
type MockQuodlibetGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockQuodlibetGame) Reset() { _m.Called() }

// NextDeal モック
func (_m *MockQuodlibetGame) NextDeal() { _m.Called() }

// SelectContract モック
func (_m *MockQuodlibetGame) SelectContract(contract int) error {
	ret := _m.Called(contract)
	if err, ok := ret.Get(0).(error); ok {
		return err
	}
	return nil
}

// CpuSelectContract モック
func (_m *MockQuodlibetGame) CpuSelectContract() { _m.Called() }

// PlayerPlay モック
func (_m *MockQuodlibetGame) PlayerPlay(handIdx int) error {
	ret := _m.Called(handIdx)
	if err, ok := ret.Get(0).(error); ok {
		return err
	}
	return nil
}

// CpuPlay モック
func (_m *MockQuodlibetGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockQuodlibetGame) GetConfig() domain.QuodlibetConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.QuodlibetConfig)
}

// SetConfig モック
func (_m *MockQuodlibetGame) SetConfig(cfg domain.QuodlibetConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockQuodlibetGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockQuodlibetGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockQuodlibetGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetDealNumber モック
func (_m *MockQuodlibetGame) GetDealNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockQuodlibetGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockQuodlibetGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentContract モック
func (_m *MockQuodlibetGame) GetCurrentContract() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAvailableContracts モック
func (_m *MockQuodlibetGame) GetAvailableContracts() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetUsedContracts モック
func (_m *MockQuodlibetGame) GetUsedContracts() [domain.QuodlibetContractCnt]bool {
	ret := _m.Called()
	return ret.Get(0).([domain.QuodlibetContractCnt]bool)
}

// GetTrickNumber モック
func (_m *MockQuodlibetGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTurn モック
func (_m *MockQuodlibetGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLeadPlayerIdx モック
func (_m *MockQuodlibetGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockQuodlibetGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLastTrick モック
func (_m *MockQuodlibetGame) GetLastTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLastTrickWinner モック
func (_m *MockQuodlibetGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayableIndices モック
func (_m *MockQuodlibetGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetSheddingPlayableIndices モック
func (_m *MockQuodlibetGame) GetSheddingPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetTablePlaced モック
func (_m *MockQuodlibetGame) GetTablePlaced() [5]uint16 {
	ret := _m.Called()
	return ret.Get(0).([5]uint16)
}

// GetStack モック
func (_m *MockQuodlibetGame) GetStack() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockQuodlibetGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockQuodlibetGame) GetPlayer(i int) *domain.QuodlibetPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.QuodlibetPlayer); ok {
		return v
	}
	return nil
}

// GetLastDealDetail モック
func (_m *MockQuodlibetGame) GetLastDealDetail() *domain.QuodlibetDealDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.QuodlibetDealDetail); ok {
		return v
	}
	return nil
}

// GetDealHistory モック
func (_m *MockQuodlibetGame) GetDealHistory() []*domain.QuodlibetDealDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.QuodlibetDealDetail); ok {
		return v
	}
	return nil
}

// GetWinners モック
func (_m *MockQuodlibetGame) GetWinners() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockQuodlibetGame) GetHint() *domain.QuodlibetHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.QuodlibetHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockQuodlibetGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBarbuGame はバルブゲームのモック (testify/mock)。
type MockBarbuGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockBarbuGame) Reset() { _m.Called() }

// NextDeal モック
func (_m *MockBarbuGame) NextDeal() { _m.Called() }

// SelectContract モック
func (_m *MockBarbuGame) SelectContract(contract, trumpSuit int) error {
	ret := _m.Called(contract, trumpSuit)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockBarbuGame) PlayerPlay(handIdx int, tableIdxs []int) error {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockBarbuGame) CpuPlay() { _m.Called() }

// SetConfig モック
func (_m *MockBarbuGame) SetConfig(config domain.BarbuConfig) { _m.Called(config) }

// GetGameEndFlag モック
func (_m *MockBarbuGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockBarbuGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerCnt モック
func (_m *MockBarbuGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockBarbuGame) GetPlayer(i int) *domain.BarbuPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.BarbuPlayer); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockBarbuGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDealerIdx モック
func (_m *MockBarbuGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDealNumber モック
func (_m *MockBarbuGame) GetDealNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentContract モック
func (_m *MockBarbuGame) GetCurrentContract() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrumpSuit モック
func (_m *MockBarbuGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrickNumber モック
func (_m *MockBarbuGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayableIndices モック
func (_m *MockBarbuGame) GetPlayableIndices(playerIdx int) []int {
	out, _ := _m.Called(playerIdx).Get(0).([]int)
	return out
}

// GetCurrentTrick モック
func (_m *MockBarbuGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLastTrick モック
func (_m *MockBarbuGame) GetLastTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLastTrickWinner モック
func (_m *MockBarbuGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTablePlaced モック
func (_m *MockBarbuGame) GetTablePlaced() [5]uint16 {
	ret := _m.Called()
	if v, ok := ret.Get(0).([5]uint16); ok {
		return v
	}
	return [5]uint16{}
}

// GetUsedContracts モック
func (_m *MockBarbuGame) GetUsedContracts(dealerIdx int) [domain.BarbuContractCnt]bool {
	ret := _m.Called(dealerIdx)
	if v, ok := ret.Get(0).([domain.BarbuContractCnt]bool); ok {
		return v
	}
	return [domain.BarbuContractCnt]bool{}
}

// GetDominoPlayableIndices モック
func (_m *MockBarbuGame) GetDominoPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetLastDealDetail モック
func (_m *MockBarbuGame) GetLastDealDetail() *domain.BarbuDealDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.BarbuDealDetail); ok {
		return v
	}
	return nil
}

// GetDealHistory モック
func (_m *MockBarbuGame) GetDealHistory() []*domain.BarbuDealDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.BarbuDealDetail); ok {
		return v
	}
	return nil
}

// GetConfig モック
func (_m *MockBarbuGame) GetConfig() domain.BarbuConfig {
	ret := _m.Called()
	if v, ok := ret.Get(0).(domain.BarbuConfig); ok {
		return v
	}
	return domain.BarbuConfig{}
}

// GetPhase モック
func (_m *MockBarbuGame) GetPhase() string {
	ret := _m.Called()
	return ret.String(0)
}

// GetRoundWinners モック
func (_m *MockBarbuGame) GetRoundWinners() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockBarbuGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEscobaGame エスコバゲームのモック (testify/mock)。
type MockEscobaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockEscobaGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockEscobaGame) NextRound() { _m.Called() }

// GetGameEndFlag モック
func (_m *MockEscobaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockEscobaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockEscobaGame) PlayerPlay(handIdx int, tableIdxs []int) error {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockEscobaGame) CpuPlay() { _m.Called() }

// SetConfig モック
func (_m *MockEscobaGame) SetConfig(config domain.EscobaConfig) { _m.Called(config) }

// GetPlayerCnt モック
func (_m *MockEscobaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockEscobaGame) GetPlayer(i int) *domain.ScopaPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.ScopaPlayer); ok {
		return v
	}
	return nil
}

// GetTableCards モック
func (_m *MockEscobaGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetStockRemaining モック
func (_m *MockEscobaGame) GetStockRemaining() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastCaptureIdx モック
func (_m *MockEscobaGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockEscobaGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDealerIdx モック
func (_m *MockEscobaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundNumber モック
func (_m *MockEscobaGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetWinnerIdx モック
func (_m *MockEscobaGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockEscobaGame) GetConfig() domain.EscobaConfig {
	ret := _m.Called()
	if v, ok := ret.Get(0).(domain.EscobaConfig); ok {
		return v
	}
	return domain.EscobaConfig{}
}

// GetPhase モック
func (_m *MockEscobaGame) GetPhase() string {
	ret := _m.Called()
	return ret.String(0)
}

// GetLastRoundDetail モック
func (_m *MockEscobaGame) GetLastRoundDetail() *domain.EscobaScoreDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.EscobaScoreDetail); ok {
		return v
	}
	return nil
}

// GetValidCaptures モック
func (_m *MockEscobaGame) GetValidCaptures(handIdx int) [][]int {
	ret := _m.Called(handIdx)
	if v, ok := ret.Get(0).([][]int); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockEscobaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

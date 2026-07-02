//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBasraGame はバスラ (Basra) のゲームモック。
type MockBasraGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockBasraGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockBasraGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockBasraGame) PlayerPlay(handIdx int, tableIdxs []int) error {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockBasraGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockBasraGame) GetConfig() domain.BasraConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BasraConfig)
}

// SetConfig モック
func (_m *MockBasraGame) SetConfig(cfg domain.BasraConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockBasraGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockBasraGame) GetPhase() domain.BasraPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BasraPhase)
}

// IsHumanTurn モック
func (_m *MockBasraGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCurrentTurn モック
func (_m *MockBasraGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTableCards モック
func (_m *MockBasraGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetLastCaptureIdx モック
func (_m *MockBasraGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRemainingDeck モック
func (_m *MockBasraGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockBasraGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinners モック
func (_m *MockBasraGame) GetWinners() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetLastDealDetail モック
func (_m *MockBasraGame) GetLastDealDetail() *domain.BasraScoreDetail {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.BasraScoreDetail)
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockBasraGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockBasraGame) GetPlayer(i int) *domain.BasraPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.BasraPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockBasraGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetCaptureOptions モック
func (_m *MockBasraGame) GetCaptureOptions(playerIdx int) map[int][]int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.(map[int][]int)
	}
	return nil
}

// GetHint モック
func (_m *MockBasraGame) GetHint() *domain.BasraHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.BasraHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockBasraGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

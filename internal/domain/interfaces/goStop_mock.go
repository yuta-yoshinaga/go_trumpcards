//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGoStopGame はゴーストップ (Go-Stop) のゲームモック。
type MockGoStopGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockGoStopGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockGoStopGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockGoStopGame) PlayerPlay(handIdx, fieldIdx int) error {
	ret := _m.Called(handIdx, fieldIdx)
	return ret.Error(0)
}

// PlayerDecide モック
func (_m *MockGoStopGame) PlayerDecide(goDecision bool) error {
	ret := _m.Called(goDecision)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockGoStopGame) CpuPlay() { _m.Called() }

// CpuDecide モック
func (_m *MockGoStopGame) CpuDecide() { _m.Called() }

// GetConfig モック
func (_m *MockGoStopGame) GetConfig() domain.GoStopConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GoStopConfig)
}

// SetConfig モック
func (_m *MockGoStopGame) SetConfig(cfg domain.GoStopConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockGoStopGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockGoStopGame) GetPhase() domain.GoStopPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GoStopPhase)
}

// IsHumanTurn モック
func (_m *MockGoStopGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCurrentTurn モック
func (_m *MockGoStopGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetFieldCards モック
func (_m *MockGoStopGame) GetFieldCards() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetRemainingDeck モック
func (_m *MockGoStopGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockGoStopGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundWinner モック
func (_m *MockGoStopGame) GetRoundWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastRoundResult モック
func (_m *MockGoStopGame) GetLastRoundResult() *domain.GoStopRoundResult {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.GoStopRoundResult)
	}
	return nil
}

// GetPendingBreakdown モック
func (_m *MockGoStopGame) GetPendingBreakdown() *domain.GoStopBreakdown {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.GoStopBreakdown)
	}
	return nil
}

// GetPendingPoints モック
func (_m *MockGoStopGame) GetPendingPoints() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinner モック
func (_m *MockGoStopGame) GetWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetResult モック
func (_m *MockGoStopGame) GetResult() domain.GoStopResult {
	ret := _m.Called()
	return ret.Get(0).(domain.GoStopResult)
}

// GetPlayerCnt モック
func (_m *MockGoStopGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockGoStopGame) GetPlayer(i int) *domain.GoStopPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.GoStopPlayer)
	}
	return nil
}

// GetScore モック
func (_m *MockGoStopGame) GetScore(playerIdx int) (*domain.GoStopBreakdown, int) {
	ret := _m.Called(playerIdx)
	var bd *domain.GoStopBreakdown
	if v := ret.Get(0); v != nil {
		bd = v.(*domain.GoStopBreakdown)
	}
	return bd, ret.Int(1)
}

// GetPlayableIndices モック
func (_m *MockGoStopGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetCaptureOptions モック
func (_m *MockGoStopGame) GetCaptureOptions(playerIdx int) map[int][]int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.(map[int][]int)
	}
	return nil
}

// GetHint モック
func (_m *MockGoStopGame) GetHint() *domain.GoStopHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.GoStopHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockGoStopGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

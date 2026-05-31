//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockScopaGame スコパゲームのモック (testify/mock)。
type MockScopaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockScopaGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockScopaGame) NextRound() { _m.Called() }

// GetGameEndFlag モック
func (_m *MockScopaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockScopaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockScopaGame) PlayerPlay(handIdx int, tableIdxs []int) error {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockScopaGame) CpuPlay() { _m.Called() }

// SetConfig モック
func (_m *MockScopaGame) SetConfig(config domain.ScopaConfig) { _m.Called(config) }

// GetPlayerCnt モック
func (_m *MockScopaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockScopaGame) GetPlayer(i int) *domain.ScopaPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.ScopaPlayer); ok {
		return v
	}
	return nil
}

// GetTableCards モック
func (_m *MockScopaGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetLastCaptureIdx モック
func (_m *MockScopaGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanAction モック
func (_m *MockScopaGame) GetHumanAction() *domain.ScopaAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.ScopaAction); ok {
		return v
	}
	return nil
}

// GetCpuActions モック
func (_m *MockScopaGame) GetCpuActions() []*domain.ScopaAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ScopaAction); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockScopaGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockScopaGame) GetConfig() domain.ScopaConfig {
	ret := _m.Called()
	if v, ok := ret.Get(0).(domain.ScopaConfig); ok {
		return v
	}
	return domain.ScopaConfig{}
}

// GetPhase モック
func (_m *MockScopaGame) GetPhase() string {
	ret := _m.Called()
	return ret.String(0)
}

// GetLastRoundDetail モック
func (_m *MockScopaGame) GetLastRoundDetail() *domain.ScopaScoreDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.ScopaScoreDetail); ok {
		return v
	}
	return nil
}

// GetRoundWinners モック
func (_m *MockScopaGame) GetRoundWinners() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetRemainingDeck モック
func (_m *MockScopaGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPacksDealt モック
func (_m *MockScopaGame) GetPacksDealt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockScopaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

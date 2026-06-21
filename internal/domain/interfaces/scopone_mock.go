//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockScoponeGame スコポーネゲームのモック (testify/mock)。
type MockScoponeGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockScoponeGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockScoponeGame) NextRound() { _m.Called() }

// GetGameEndFlag モック
func (_m *MockScoponeGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockScoponeGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockScoponeGame) PlayerPlay(handIdx int, tableIdxs []int) error {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockScoponeGame) CpuPlay() { _m.Called() }

// SetConfig モック
func (_m *MockScoponeGame) SetConfig(config domain.ScoponeConfig) { _m.Called(config) }

// GetPlayerCnt モック
func (_m *MockScoponeGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockScoponeGame) GetPlayer(i int) *domain.ScopaPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.ScopaPlayer); ok {
		return v
	}
	return nil
}

// GetTableCards モック
func (_m *MockScoponeGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetLastCaptureIdx モック
func (_m *MockScoponeGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockScoponeGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDealerIdx モック
func (_m *MockScoponeGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundNumber モック
func (_m *MockScoponeGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTeamScore モック
func (_m *MockScoponeGame) GetTeamScore(team int) int {
	ret := _m.Called(team)
	return ret.Int(0)
}

// GetWinnerTeam モック
func (_m *MockScoponeGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockScoponeGame) GetConfig() domain.ScoponeConfig {
	ret := _m.Called()
	if v, ok := ret.Get(0).(domain.ScoponeConfig); ok {
		return v
	}
	return domain.ScoponeConfig{}
}

// GetPhase モック
func (_m *MockScoponeGame) GetPhase() string {
	ret := _m.Called()
	return ret.String(0)
}

// GetLastRoundDetail モック
func (_m *MockScoponeGame) GetLastRoundDetail() *domain.ScoponeScoreDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.ScoponeScoreDetail); ok {
		return v
	}
	return nil
}

// GetValidCaptures モック
func (_m *MockScoponeGame) GetValidCaptures(handIdx int) [][]int {
	ret := _m.Called(handIdx)
	if v, ok := ret.Get(0).([][]int); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockScoponeGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

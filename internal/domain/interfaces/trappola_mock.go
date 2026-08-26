//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrappolaGame トラッポラのゲームモック
type MockTrappolaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTrappolaGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockTrappolaGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockTrappolaGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockTrappolaGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockTrappolaGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockTrappolaGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockTrappolaGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockTrappolaGame) GetConfig() domain.TrappolaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TrappolaConfig)
}

// SetConfig モック
func (_m *MockTrappolaGame) SetConfig(cfg domain.TrappolaConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockTrappolaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockTrappolaGame) GetPhase() domain.TrappolaPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TrappolaPhase)
}

// IsHumanTurn モック
func (_m *MockTrappolaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockTrappolaGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockTrappolaGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockTrappolaGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockTrappolaGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	return ret.Get(0).([]*domain.TrickCard)
}

// GetLeadPlayerIdx モック
func (_m *MockTrappolaGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamScores モック
func (_m *MockTrappolaGame) GetTeamScores() [domain.TrappolaTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TrappolaTeamCnt]int)
}

// GetTeamRoundThirds モック
func (_m *MockTrappolaGame) GetTeamRoundThirds() [domain.TrappolaTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TrappolaTeamCnt]int)
}

// GetWinnerTeam モック
func (_m *MockTrappolaGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockTrappolaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockTrappolaGame) GetPlayer(i int) *domain.TrappolaPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.TrappolaPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockTrappolaGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockTrappolaGame) GetHint() *domain.TrappolaHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TrappolaHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockTrappolaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

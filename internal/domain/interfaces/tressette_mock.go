//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTressetteGame トレセッテのゲームモック
type MockTressetteGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTressetteGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockTressetteGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockTressetteGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockTressetteGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockTressetteGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockTressetteGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockTressetteGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockTressetteGame) GetConfig() domain.TressetteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TressetteConfig)
}

// SetConfig モック
func (_m *MockTressetteGame) SetConfig(cfg domain.TressetteConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockTressetteGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockTressetteGame) GetPhase() domain.TressettePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TressettePhase)
}

// IsHumanTurn モック
func (_m *MockTressetteGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockTressetteGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockTressetteGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockTressetteGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockTressetteGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	return ret.Get(0).([]*domain.TrickCard)
}

// GetLeadPlayerIdx モック
func (_m *MockTressetteGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamScores モック
func (_m *MockTressetteGame) GetTeamScores() [domain.TressetteTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TressetteTeamCnt]int)
}

// GetTeamRoundThirds モック
func (_m *MockTressetteGame) GetTeamRoundThirds() [domain.TressetteTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TressetteTeamCnt]int)
}

// GetWinnerTeam モック
func (_m *MockTressetteGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockTressetteGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockTressetteGame) GetPlayer(i int) *domain.TressettePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.TressettePlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockTressetteGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockTressetteGame) GetHint() *domain.TressetteHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TressetteHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockTressetteGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

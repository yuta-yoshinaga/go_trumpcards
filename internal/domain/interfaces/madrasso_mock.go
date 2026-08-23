//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMadrassoGame マドラッソのゲームモック
type MockMadrassoGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockMadrassoGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockMadrassoGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockMadrassoGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockMadrassoGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockMadrassoGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockMadrassoGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockMadrassoGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockMadrassoGame) GetConfig() domain.MadrassoConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MadrassoConfig)
}

// SetConfig モック
func (_m *MockMadrassoGame) SetConfig(cfg domain.MadrassoConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockMadrassoGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockMadrassoGame) GetPhase() domain.MadrassoPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MadrassoPhase)
}

// IsHumanTurn モック
func (_m *MockMadrassoGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockMadrassoGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockMadrassoGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockMadrassoGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockMadrassoGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	return ret.Get(0).([]*domain.TrickCard)
}

// GetLeadPlayerIdx モック
func (_m *MockMadrassoGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamScores モック
func (_m *MockMadrassoGame) GetTeamScores() [domain.MadrassoTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MadrassoTeamCnt]int)
}

// GetTrumpSuit モック
func (_m *MockMadrassoGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamRoundPoints モック
func (_m *MockMadrassoGame) GetTeamRoundPoints() [domain.MadrassoTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MadrassoTeamCnt]int)
}

// GetWinnerTeam モック
func (_m *MockMadrassoGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockMadrassoGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockMadrassoGame) GetPlayer(i int) *domain.MadrassoPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.MadrassoPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockMadrassoGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockMadrassoGame) GetHint() *domain.MadrassoHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.MadrassoHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockMadrassoGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

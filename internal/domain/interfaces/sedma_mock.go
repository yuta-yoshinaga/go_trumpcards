//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSedmaGame セドマのゲームモック
type MockSedmaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSedmaGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockSedmaGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockSedmaGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSedmaGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockSedmaGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockSedmaGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockSedmaGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockSedmaGame) GetConfig() domain.SedmaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SedmaConfig)
}

// SetConfig モック
func (_m *MockSedmaGame) SetConfig(cfg domain.SedmaConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockSedmaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockSedmaGame) GetPhase() domain.SedmaPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SedmaPhase)
}

// IsHumanTurn モック
func (_m *MockSedmaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockSedmaGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockSedmaGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockSedmaGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockSedmaGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockSedmaGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockSedmaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamScores モック
func (_m *MockSedmaGame) GetTeamScores() [domain.SedmaTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.SedmaTeamCnt]int)
}

// GetRoundCardPoints モック
func (_m *MockSedmaGame) GetRoundCardPoints() [domain.SedmaTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.SedmaTeamCnt]int)
}

// GetWinnerTeam モック
func (_m *MockSedmaGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockSedmaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSedmaGame) GetPlayer(i int) *domain.SedmaPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SedmaPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockSedmaGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockSedmaGame) GetHint() *domain.SedmaHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.SedmaHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockSedmaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

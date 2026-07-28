//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKlaverjasGame クラヴァヤスのゲームモック
type MockKlaverjasGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockKlaverjasGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockKlaverjasGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockKlaverjasGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockKlaverjasGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockKlaverjasGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockKlaverjasGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockKlaverjasGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockKlaverjasGame) GetConfig() domain.KlaverjasConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KlaverjasConfig)
}

// SetConfig モック
func (_m *MockKlaverjasGame) SetConfig(cfg domain.KlaverjasConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockKlaverjasGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockKlaverjasGame) GetPhase() domain.KlaverjasPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.KlaverjasPhase)
}

// IsHumanTurn モック
func (_m *MockKlaverjasGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockKlaverjasGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockKlaverjasGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockKlaverjasGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockKlaverjasGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockKlaverjasGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockKlaverjasGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockKlaverjasGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamScores モック
func (_m *MockKlaverjasGame) GetTeamScores() [domain.KlaverjasTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.KlaverjasTeamCnt]int)
}

// GetRoundCardPoints モック
func (_m *MockKlaverjasGame) GetRoundCardPoints() [domain.KlaverjasTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.KlaverjasTeamCnt]int)
}

// GetRoundRoem モック
func (_m *MockKlaverjasGame) GetRoundRoem() [domain.KlaverjasTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.KlaverjasTeamCnt]int)
}

// GetWinnerTeam モック
func (_m *MockKlaverjasGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockKlaverjasGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockKlaverjasGame) GetPlayer(i int) *domain.KlaverjasPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.KlaverjasPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockKlaverjasGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockKlaverjasGame) GetHint() *domain.KlaverjasHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.KlaverjasHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockKlaverjasGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

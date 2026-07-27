//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockManilleGame マニーユのゲームモック
type MockManilleGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockManilleGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockManilleGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockManilleGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockManilleGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockManilleGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockManilleGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockManilleGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockManilleGame) GetConfig() domain.ManilleConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ManilleConfig)
}

// SetConfig モック
func (_m *MockManilleGame) SetConfig(cfg domain.ManilleConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockManilleGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockManilleGame) GetPhase() domain.ManillePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.ManillePhase)
}

// IsHumanTurn モック
func (_m *MockManilleGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockManilleGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockManilleGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockManilleGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockManilleGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockManilleGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockManilleGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockManilleGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamScores モック
func (_m *MockManilleGame) GetTeamScores() [domain.ManilleTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.ManilleTeamCnt]int)
}

// GetRoundCardPoints モック
func (_m *MockManilleGame) GetRoundCardPoints() [domain.ManilleTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.ManilleTeamCnt]int)
}

// GetWinnerTeam モック
func (_m *MockManilleGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockManilleGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockManilleGame) GetPlayer(i int) *domain.ManillePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.ManillePlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockManilleGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockManilleGame) GetHint() *domain.ManilleHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.ManilleHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockManilleGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTuteGame トゥーテのゲームモック
type MockTuteGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTuteGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockTuteGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockTuteGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// PlayerDeclareMarriage モック
func (_m *MockTuteGame) PlayerDeclareMarriage(suit int) error {
	ret := _m.Called(suit)
	return ret.Error(0)
}

// PlayerDeclareTute モック
func (_m *MockTuteGame) PlayerDeclareTute() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockTuteGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockTuteGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockTuteGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockTuteGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockTuteGame) GetConfig() domain.TuteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TuteConfig)
}

// SetConfig モック
func (_m *MockTuteGame) SetConfig(cfg domain.TuteConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockTuteGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockTuteGame) GetPhase() domain.TutePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TutePhase)
}

// IsHumanTurn モック
func (_m *MockTuteGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockTuteGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockTuteGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockTuteGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockTuteGame) GetCurrentTrick() []*domain.TuteTrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TuteTrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockTuteGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockTuteGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockTuteGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsSuitDeclared モック
func (_m *MockTuteGame) IsSuitDeclared(suit int) bool {
	ret := _m.Called(suit)
	return ret.Get(0).(bool)
}

// GetTeamScores モック
func (_m *MockTuteGame) GetTeamScores() [domain.TuteTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TuteTeamCnt]int)
}

// GetRoundTeamPoints モック
func (_m *MockTuteGame) GetRoundTeamPoints() [domain.TuteTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TuteTeamCnt]int)
}

// GetWinnerTeam モック
func (_m *MockTuteGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockTuteGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockTuteGame) GetPlayer(i int) *domain.TutePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.TutePlayer)
	}
	return nil
}

// CanHumanDeclareMarriage モック
func (_m *MockTuteGame) CanHumanDeclareMarriage() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetHumanDeclarableMarriageSuits モック
func (_m *MockTuteGame) GetHumanDeclarableMarriageSuits() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// CanHumanDeclareTute モック
func (_m *MockTuteGame) CanHumanDeclareTute() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPlayableIndices モック
func (_m *MockTuteGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockTuteGame) GetHint() *domain.TuteHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TuteHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockTuteGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

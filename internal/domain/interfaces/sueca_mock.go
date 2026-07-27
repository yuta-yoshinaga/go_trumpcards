//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSuecaGame スエカのゲームモック
type MockSuecaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSuecaGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockSuecaGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockSuecaGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSuecaGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockSuecaGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockSuecaGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockSuecaGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockSuecaGame) GetConfig() domain.SuecaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SuecaConfig)
}

// SetConfig モック
func (_m *MockSuecaGame) SetConfig(cfg domain.SuecaConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockSuecaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockSuecaGame) GetPhase() domain.SuecaPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SuecaPhase)
}

// IsHumanTurn モック
func (_m *MockSuecaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockSuecaGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockSuecaGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockSuecaGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockSuecaGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockSuecaGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockSuecaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockSuecaGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamGamePoints モック
func (_m *MockSuecaGame) GetTeamGamePoints() [domain.SuecaTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.SuecaTeamCnt]int)
}

// GetRoundCardPoints モック
func (_m *MockSuecaGame) GetRoundCardPoints() [domain.SuecaTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.SuecaTeamCnt]int)
}

// GetRoundWinnerTeam モック
func (_m *MockSuecaGame) GetRoundWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundGamePoints モック
func (_m *MockSuecaGame) GetRoundGamePoints() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerTeam モック
func (_m *MockSuecaGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockSuecaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSuecaGame) GetPlayer(i int) *domain.SuecaPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SuecaPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockSuecaGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockSuecaGame) GetHint() *domain.SuecaHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.SuecaHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockSuecaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

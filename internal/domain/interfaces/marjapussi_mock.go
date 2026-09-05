//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMarjapussiGame マルヤプッシ (Marjapussi) のゲームモック
type MockMarjapussiGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockMarjapussiGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockMarjapussiGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockMarjapussiGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockMarjapussiGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockMarjapussiGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockMarjapussiGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockMarjapussiGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockMarjapussiGame) GetConfig() domain.MarjapussiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MarjapussiConfig)
}

// SetConfig モック
func (_m *MockMarjapussiGame) SetConfig(cfg domain.MarjapussiConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockMarjapussiGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockMarjapussiGame) GetPhase() domain.MarjapussiPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MarjapussiPhase)
}

// IsHumanTurn モック
func (_m *MockMarjapussiGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockMarjapussiGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockMarjapussiGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockMarjapussiGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockMarjapussiGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockMarjapussiGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockMarjapussiGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockMarjapussiGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPussi モック
func (_m *MockMarjapussiGame) GetPussi() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetTeamScores モック
func (_m *MockMarjapussiGame) GetTeamScores() [domain.MarjapussiTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MarjapussiTeamCnt]int)
}

// GetPlayerScores モック
func (_m *MockMarjapussiGame) GetPlayerScores() [domain.MarjapussiPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MarjapussiPlayerCnt]int)
}

// GetRoundCardPoints モック
func (_m *MockMarjapussiGame) GetRoundCardPoints() [domain.MarjapussiTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MarjapussiTeamCnt]int)
}

// GetRoundMarriage モック
func (_m *MockMarjapussiGame) GetRoundMarriage() [domain.MarjapussiTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MarjapussiTeamCnt]int)
}

// GetMarriageOptions モック
func (_m *MockMarjapussiGame) GetMarriageOptions(playerIdx int) []domain.MarjapussiMarriageOption {
	ret := _m.Called(playerIdx)
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]domain.MarjapussiMarriageOption)
}

// GetWinnerPlayer モック
func (_m *MockMarjapussiGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerTeam モック
func (_m *MockMarjapussiGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockMarjapussiGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockMarjapussiGame) GetPlayer(i int) *domain.MarjapussiPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.MarjapussiPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockMarjapussiGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockMarjapussiGame) GetHint() *domain.MarjapussiHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.MarjapussiHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockMarjapussiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

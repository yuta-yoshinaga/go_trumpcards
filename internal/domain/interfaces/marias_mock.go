//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMariasGame マリアーシュのゲームモック
type MockMariasGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockMariasGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockMariasGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockMariasGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockMariasGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockMariasGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockMariasGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockMariasGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockMariasGame) GetConfig() domain.MariasConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MariasConfig)
}

// SetConfig モック
func (_m *MockMariasGame) SetConfig(cfg domain.MariasConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockMariasGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockMariasGame) GetPhase() domain.MariasPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MariasPhase)
}

// IsHumanTurn モック
func (_m *MockMariasGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockMariasGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockMariasGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockMariasGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockMariasGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockMariasGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockMariasGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetSoloistIdx モック
func (_m *MockMariasGame) GetSoloistIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockMariasGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockMariasGame) GetPlayerScores() [domain.MariasPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MariasPlayerCnt]int)
}

// GetRoundCardPoints モック
func (_m *MockMariasGame) GetRoundCardPoints() [domain.MariasPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MariasPlayerCnt]int)
}

// GetRoundMarriage モック
func (_m *MockMariasGame) GetRoundMarriage() [domain.MariasPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MariasPlayerCnt]int)
}

// GetWinnerPlayer モック
func (_m *MockMariasGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockMariasGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockMariasGame) GetPlayer(i int) *domain.MariasPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.MariasPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockMariasGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockMariasGame) GetHint() *domain.MariasHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.MariasHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockMariasGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

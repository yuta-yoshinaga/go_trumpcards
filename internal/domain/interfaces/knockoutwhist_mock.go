//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKnockoutWhistGame ノックアウト・ホイストのゲームモック
type MockKnockoutWhistGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockKnockoutWhistGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockKnockoutWhistGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockKnockoutWhistGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// PlayerSelectTrump モック
func (_m *MockKnockoutWhistGame) PlayerSelectTrump(suit int) error {
	ret := _m.Called(suit)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockKnockoutWhistGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockKnockoutWhistGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockKnockoutWhistGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockKnockoutWhistGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockKnockoutWhistGame) GetConfig() domain.KnockoutWhistConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KnockoutWhistConfig)
}

// SetConfig モック
func (_m *MockKnockoutWhistGame) SetConfig(cfg domain.KnockoutWhistConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockKnockoutWhistGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockKnockoutWhistGame) GetPhase() domain.KnockoutWhistPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.KnockoutWhistPhase)
}

// IsHumanTurn モック
func (_m *MockKnockoutWhistGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockKnockoutWhistGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHandSize モック
func (_m *MockKnockoutWhistGame) GetHandSize() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockKnockoutWhistGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockKnockoutWhistGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockKnockoutWhistGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockKnockoutWhistGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockKnockoutWhistGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockKnockoutWhistGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundWinnerIdx モック
func (_m *MockKnockoutWhistGame) GetRoundWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerPlayer モック
func (_m *MockKnockoutWhistGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetActiveCount モック
func (_m *MockKnockoutWhistGame) GetActiveCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockKnockoutWhistGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockKnockoutWhistGame) GetPlayer(i int) *domain.KnockoutWhistPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.KnockoutWhistPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockKnockoutWhistGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockKnockoutWhistGame) GetHint() *domain.KnockoutWhistHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.KnockoutWhistHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockKnockoutWhistGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

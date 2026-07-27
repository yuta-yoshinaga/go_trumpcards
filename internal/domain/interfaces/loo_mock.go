//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLooGame はルー (Loo) のゲームモック。
type MockLooGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockLooGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockLooGame) NextRound() { _m.Called() }

// PlayerDecide モック
func (_m *MockLooGame) PlayerDecide(play bool) error {
	ret := _m.Called(play)
	return ret.Error(0)
}

// CpuDecide モック
func (_m *MockLooGame) CpuDecide() { _m.Called() }

// PlayerPlay モック
func (_m *MockLooGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockLooGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockLooGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockLooGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockLooGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockLooGame) GetConfig() domain.LooConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.LooConfig)
}

// SetConfig モック
func (_m *MockLooGame) SetConfig(cfg domain.LooConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockLooGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockLooGame) GetPhase() domain.LooPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.LooPhase)
}

// IsHumanTurn モック
func (_m *MockLooGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockLooGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockLooGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockLooGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTurn モック
func (_m *MockLooGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDecidePlayerIdx モック
func (_m *MockLooGame) GetDecidePlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockLooGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLastTrick モック
func (_m *MockLooGame) GetLastTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLastTrickWinner モック
func (_m *MockLooGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLeadPlayerIdx モック
func (_m *MockLooGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockLooGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTurnUp モック
func (_m *MockLooGame) GetTurnUp() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetPot モック
func (_m *MockLooGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPotStart モック
func (_m *MockLooGame) GetPotStart() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastDealDetail モック
func (_m *MockLooGame) GetLastDealDetail() *domain.LooDealDetail {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.LooDealDetail)
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockLooGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockLooGame) GetPlayer(i int) *domain.LooPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.LooPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockLooGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockLooGame) GetHint() *domain.LooHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.LooHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockLooGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCinchGame はチンチ (Cinch) のゲームモック。
type MockCinchGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockCinchGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockCinchGame) NextRound() { _m.Called() }

// PlayerBid モック
func (_m *MockCinchGame) PlayerBid(bid int) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockCinchGame) CpuBid() { _m.Called() }

// NameTrump モック
func (_m *MockCinchGame) NameTrump(suit int) error {
	ret := _m.Called(suit)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockCinchGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockCinchGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockCinchGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockCinchGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockCinchGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockCinchGame) GetConfig() domain.CinchConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CinchConfig)
}

// SetConfig モック
func (_m *MockCinchGame) SetConfig(cfg domain.CinchConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockCinchGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockCinchGame) GetPhase() domain.CinchPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CinchPhase)
}

// IsHumanTurn モック
func (_m *MockCinchGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockCinchGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockCinchGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockCinchGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTurn モック
func (_m *MockCinchGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockCinchGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLastTrick モック
func (_m *MockCinchGame) GetLastTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLastTrickWinner モック
func (_m *MockCinchGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLeadPlayerIdx モック
func (_m *MockCinchGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetBidPlayerIdx モック
func (_m *MockCinchGame) GetBidPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBid モック
func (_m *MockCinchGame) GetCurrentBid() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetBidWinnerIdx モック
func (_m *MockCinchGame) GetBidWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockCinchGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerIdx モック
func (_m *MockCinchGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastDealDetail モック
func (_m *MockCinchGame) GetLastDealDetail() *domain.CinchDealDetail {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.CinchDealDetail)
	}
	return nil
}

// GetRoundWinners モック
func (_m *MockCinchGame) GetRoundWinners() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockCinchGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockCinchGame) GetPlayer(i int) *domain.CinchPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.CinchPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockCinchGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockCinchGame) GetHint() *domain.CinchHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.CinchHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockCinchGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

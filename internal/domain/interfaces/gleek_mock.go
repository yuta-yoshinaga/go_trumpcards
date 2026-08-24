//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGleekGame グリーク (Gleek) のゲームモック
type MockGleekGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockGleekGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockGleekGame) NextRound() { _m.Called() }

// PlayerBid モック
func (_m *MockGleekGame) PlayerBid(bid int) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockGleekGame) CpuBid() { _m.Called() }

// NextBidAmount モック
func (_m *MockGleekGame) NextBidAmount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// HighestBid モック
func (_m *MockGleekGame) HighestBid() int {
	ret := _m.Called()
	return ret.Int(0)
}

// PlayerDiscard モック
func (_m *MockGleekGame) PlayerDiscard(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuDiscard モック
func (_m *MockGleekGame) CpuDiscard() { _m.Called() }

// GetDiscardHint モック
func (_m *MockGleekGame) GetDiscardHint() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// PlayerPlay モック
func (_m *MockGleekGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockGleekGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockGleekGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockGleekGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockGleekGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockGleekGame) GetConfig() domain.GleekConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GleekConfig)
}

// SetConfig モック
func (_m *MockGleekGame) SetConfig(cfg domain.GleekConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockGleekGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockGleekGame) GetPhase() domain.GleekPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GleekPhase)
}

// IsHumanTurn モック
func (_m *MockGleekGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanBidTurn モック
func (_m *MockGleekGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanDiscardTurn モック
func (_m *MockGleekGame) IsHumanDiscardTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRoundNumber モック
func (_m *MockGleekGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrickNumber モック
func (_m *MockGleekGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockGleekGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTrick モック
func (_m *MockGleekGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockGleekGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDealerIdx モック
func (_m *MockGleekGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetElderIdx モック
func (_m *MockGleekGame) GetElderIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrumpSuit モック
func (_m *MockGleekGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTurnUp モック
func (_m *MockGleekGame) GetTurnUp() *domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

// GetCurrentBidderIdx モック
func (_m *MockGleekGame) GetCurrentBidderIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetBids モック
func (_m *MockGleekGame) GetBids() [domain.GleekPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.GleekPlayerCnt]int)
}

// GetPassed モック
func (_m *MockGleekGame) GetPassed() [domain.GleekPlayerCnt]bool {
	ret := _m.Called()
	return ret.Get(0).([domain.GleekPlayerCnt]bool)
}

// GetBuyerIdx モック
func (_m *MockGleekGame) GetBuyerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetWinningBid モック
func (_m *MockGleekGame) GetWinningBid() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRuffs モック
func (_m *MockGleekGame) GetRuffs() []*domain.GleekRuff {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.GleekRuff); ok {
		return v
	}
	return nil
}

// GetRuffWinnerIdx モック
func (_m *MockGleekGame) GetRuffWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetMelds モック
func (_m *MockGleekGame) GetMelds() []*domain.GleekMeld {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.GleekMeld); ok {
		return v
	}
	return nil
}

// GetTrickPoints モック
func (_m *MockGleekGame) GetTrickPoints() [domain.GleekPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.GleekPlayerCnt]int)
}

// DealPoints モック
func (_m *MockGleekGame) DealPoints() int {
	ret := _m.Called()
	return ret.Int(0)
}

// Par モック
func (_m *MockGleekGame) Par() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayerScores モック
func (_m *MockGleekGame) GetPlayerScores() [domain.GleekPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.GleekPlayerCnt]int)
}

// GetResult モック
func (_m *MockGleekGame) GetResult() domain.GleekResult {
	ret := _m.Called()
	return ret.Get(0).(domain.GleekResult)
}

// GetWinnerPlayer モック
func (_m *MockGleekGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayerCnt モック
func (_m *MockGleekGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockGleekGame) GetPlayer(i int) *domain.GleekPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.GleekPlayer); ok {
		return v
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockGleekGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockGleekGame) GetHint() *domain.GleekHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.GleekHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockGleekGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

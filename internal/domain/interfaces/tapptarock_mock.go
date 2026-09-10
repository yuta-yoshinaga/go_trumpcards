//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTappTarockGame はタップ・タロックのゲームモック。
type MockTappTarockGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTappTarockGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockTappTarockGame) NextRound() { _m.Called() }

// PlayerBid モック
func (_m *MockTappTarockGame) PlayerBid(bid domain.TappTarockBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// PlayerPass モック
func (_m *MockTappTarockGame) PlayerPass() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockTappTarockGame) CpuBid() { _m.Called() }

// PlayerDiscard モック
func (_m *MockTappTarockGame) PlayerDiscard(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuDiscard モック
func (_m *MockTappTarockGame) CpuDiscard() { _m.Called() }

// PlayerPlayCard モック
func (_m *MockTappTarockGame) PlayerPlayCard(handIdx int) error {
	ret := _m.Called(handIdx)
	return ret.Error(0)
}

// CpuPlayCard モック
func (_m *MockTappTarockGame) CpuPlayCard() { _m.Called() }

// NextTrick モック
func (_m *MockTappTarockGame) NextTrick() { _m.Called() }

// GetConfig モック
func (_m *MockTappTarockGame) GetConfig() domain.TappTarockConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TappTarockConfig)
}

// SetConfig モック
func (_m *MockTappTarockGame) SetConfig(cfg domain.TappTarockConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockTappTarockGame) GetPhase() domain.TappTarockPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TappTarockPhase)
}

// IsHumanTurn モック
func (_m *MockTappTarockGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// HumanSeat モック
func (_m *MockTappTarockGame) HumanSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundNumber モック
func (_m *MockTappTarockGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrickNumber モック
func (_m *MockTappTarockGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockTappTarockGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTrick モック
func (_m *MockTappTarockGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockTappTarockGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetBidPlayerIdx モック
func (_m *MockTappTarockGame) GetBidPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHighestBid モック
func (_m *MockTappTarockGame) GetHighestBid() domain.TappTarockBid {
	ret := _m.Called()
	return ret.Get(0).(domain.TappTarockBid)
}

// GetDeclarerIdx モック
func (_m *MockTappTarockGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetContract モック
func (_m *MockTappTarockGame) GetContract() domain.TappTarockBid {
	ret := _m.Called()
	return ret.Get(0).(domain.TappTarockBid)
}

// GetCalledTrump モック
func (_m *MockTappTarockGame) GetCalledTrump() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPartnerRevealed モック
func (_m *MockTappTarockGame) GetPartnerRevealed() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPartnerIdx モック
func (_m *MockTappTarockGame) GetPartnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTalonSize モック
func (_m *MockTappTarockGame) GetTalonSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastTrickWinner モック
func (_m *MockTappTarockGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastTrickCards モック
func (_m *MockTappTarockGame) GetLastTrickCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetOutcome モック
func (_m *MockTappTarockGame) GetOutcome() domain.TappTarockOutcome {
	ret := _m.Called()
	return ret.Get(0).(domain.TappTarockOutcome)
}

// GetBreakdown モック
func (_m *MockTappTarockGame) GetBreakdown() *domain.TappTarockBreakdown {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.TappTarockBreakdown); ok {
		return v
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockTappTarockGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockTappTarockGame) GetPlayer(i int) *domain.TappTarockPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.TappTarockPlayer); ok {
		return v
	}
	return nil
}

// GetPlayerScore モック
func (_m *MockTappTarockGame) GetPlayerScore(i int) int {
	ret := _m.Called(i)
	return ret.Int(0)
}

// GetCardPoints モック
func (_m *MockTappTarockGame) GetCardPoints(i int) int {
	ret := _m.Called(i)
	return ret.Int(0)
}

// GetDiscardableIndices モック
func (_m *MockTappTarockGame) GetDiscardableIndices() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetValidPlayIndices モック
func (_m *MockTappTarockGame) GetValidPlayIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetGameEndFlag モック
func (_m *MockTappTarockGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetWinnerPlayer モック
func (_m *MockTappTarockGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHint モック
func (_m *MockTappTarockGame) GetHint() *domain.TappTarockHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.TappTarockHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockTappTarockGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

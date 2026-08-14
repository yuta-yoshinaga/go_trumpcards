//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockZwanzigerrufenGame はツヴァンツィガールーフェンのゲームモック。
type MockZwanzigerrufenGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockZwanzigerrufenGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockZwanzigerrufenGame) NextRound() { _m.Called() }

// PlayerBid モック
func (_m *MockZwanzigerrufenGame) PlayerBid(bid domain.ZwanzigerrufenBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// PlayerPass モック
func (_m *MockZwanzigerrufenGame) PlayerPass() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockZwanzigerrufenGame) CpuBid() { _m.Called() }

// PlayerDiscard モック
func (_m *MockZwanzigerrufenGame) PlayerDiscard(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuDiscard モック
func (_m *MockZwanzigerrufenGame) CpuDiscard() { _m.Called() }

// PlayerPlayCard モック
func (_m *MockZwanzigerrufenGame) PlayerPlayCard(handIdx int) error {
	ret := _m.Called(handIdx)
	return ret.Error(0)
}

// CpuPlayCard モック
func (_m *MockZwanzigerrufenGame) CpuPlayCard() { _m.Called() }

// NextTrick モック
func (_m *MockZwanzigerrufenGame) NextTrick() { _m.Called() }

// GetConfig モック
func (_m *MockZwanzigerrufenGame) GetConfig() domain.ZwanzigerrufenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ZwanzigerrufenConfig)
}

// SetConfig モック
func (_m *MockZwanzigerrufenGame) SetConfig(cfg domain.ZwanzigerrufenConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockZwanzigerrufenGame) GetPhase() domain.ZwanzigerrufenPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.ZwanzigerrufenPhase)
}

// IsHumanTurn モック
func (_m *MockZwanzigerrufenGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// HumanSeat モック
func (_m *MockZwanzigerrufenGame) HumanSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundNumber モック
func (_m *MockZwanzigerrufenGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrickNumber モック
func (_m *MockZwanzigerrufenGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockZwanzigerrufenGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTrick モック
func (_m *MockZwanzigerrufenGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockZwanzigerrufenGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetBidPlayerIdx モック
func (_m *MockZwanzigerrufenGame) GetBidPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHighestBid モック
func (_m *MockZwanzigerrufenGame) GetHighestBid() domain.ZwanzigerrufenBid {
	ret := _m.Called()
	return ret.Get(0).(domain.ZwanzigerrufenBid)
}

// GetDeclarerIdx モック
func (_m *MockZwanzigerrufenGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetContract モック
func (_m *MockZwanzigerrufenGame) GetContract() domain.ZwanzigerrufenBid {
	ret := _m.Called()
	return ret.Get(0).(domain.ZwanzigerrufenBid)
}

// GetCalledTrump モック
func (_m *MockZwanzigerrufenGame) GetCalledTrump() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPartnerRevealed モック
func (_m *MockZwanzigerrufenGame) GetPartnerRevealed() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPartnerIdx モック
func (_m *MockZwanzigerrufenGame) GetPartnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTalonSize モック
func (_m *MockZwanzigerrufenGame) GetTalonSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastTrickWinner モック
func (_m *MockZwanzigerrufenGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastTrickCards モック
func (_m *MockZwanzigerrufenGame) GetLastTrickCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetOutcome モック
func (_m *MockZwanzigerrufenGame) GetOutcome() domain.ZwanzigerrufenOutcome {
	ret := _m.Called()
	return ret.Get(0).(domain.ZwanzigerrufenOutcome)
}

// GetBreakdown モック
func (_m *MockZwanzigerrufenGame) GetBreakdown() *domain.ZwanzigerrufenBreakdown {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.ZwanzigerrufenBreakdown); ok {
		return v
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockZwanzigerrufenGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockZwanzigerrufenGame) GetPlayer(i int) *domain.ZwanzigerrufenPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.ZwanzigerrufenPlayer); ok {
		return v
	}
	return nil
}

// GetPlayerScore モック
func (_m *MockZwanzigerrufenGame) GetPlayerScore(i int) int {
	ret := _m.Called(i)
	return ret.Int(0)
}

// GetCardPoints モック
func (_m *MockZwanzigerrufenGame) GetCardPoints(i int) int {
	ret := _m.Called(i)
	return ret.Int(0)
}

// GetValidPlayIndices モック
func (_m *MockZwanzigerrufenGame) GetValidPlayIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetGameEndFlag モック
func (_m *MockZwanzigerrufenGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetWinnerPlayer モック
func (_m *MockZwanzigerrufenGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHint モック
func (_m *MockZwanzigerrufenGame) GetHint() *domain.ZwanzigerrufenHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.ZwanzigerrufenHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockZwanzigerrufenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

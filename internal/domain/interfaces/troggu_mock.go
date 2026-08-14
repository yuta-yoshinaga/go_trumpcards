//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrogguGame はトロッグのゲームモック。
type MockTrogguGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTrogguGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockTrogguGame) NextRound() { _m.Called() }

// PlayerBid モック
func (_m *MockTrogguGame) PlayerBid(bid domain.TrogguBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// PlayerPass モック
func (_m *MockTrogguGame) PlayerPass() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockTrogguGame) CpuBid() { _m.Called() }

// PlayerPlayCard モック
func (_m *MockTrogguGame) PlayerPlayCard(handIdx int) error {
	ret := _m.Called(handIdx)
	return ret.Error(0)
}

// CpuPlayCard モック
func (_m *MockTrogguGame) CpuPlayCard() { _m.Called() }

// NextTrick モック
func (_m *MockTrogguGame) NextTrick() { _m.Called() }

// GetConfig モック
func (_m *MockTrogguGame) GetConfig() domain.TrogguConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TrogguConfig)
}

// SetConfig モック
func (_m *MockTrogguGame) SetConfig(cfg domain.TrogguConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockTrogguGame) GetPhase() domain.TrogguPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TrogguPhase)
}

// IsHumanTurn モック
func (_m *MockTrogguGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// HumanSeat モック
func (_m *MockTrogguGame) HumanSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundNumber モック
func (_m *MockTrogguGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrickNumber モック
func (_m *MockTrogguGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockTrogguGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTrick モック
func (_m *MockTrogguGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetDealerIdx モック
func (_m *MockTrogguGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetBidPlayerIdx モック
func (_m *MockTrogguGame) GetBidPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHighestBid モック
func (_m *MockTrogguGame) GetHighestBid() domain.TrogguBid {
	ret := _m.Called()
	return ret.Get(0).(domain.TrogguBid)
}

// GetDeclarerIdx モック
func (_m *MockTrogguGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetContract モック
func (_m *MockTrogguGame) GetContract() domain.TrogguBid {
	ret := _m.Called()
	return ret.Get(0).(domain.TrogguBid)
}

// GetTalonSize モック
func (_m *MockTrogguGame) GetTalonSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastTrickWinner モック
func (_m *MockTrogguGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastTrickCards モック
func (_m *MockTrogguGame) GetLastTrickCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetOutcome モック
func (_m *MockTrogguGame) GetOutcome() domain.TrogguOutcome {
	ret := _m.Called()
	return ret.Get(0).(domain.TrogguOutcome)
}

// GetBreakdown モック
func (_m *MockTrogguGame) GetBreakdown() *domain.TrogguBreakdown {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.TrogguBreakdown); ok {
		return v
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockTrogguGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockTrogguGame) GetPlayer(i int) *domain.TrogguPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.TrogguPlayer); ok {
		return v
	}
	return nil
}

// GetPlayerScore モック
func (_m *MockTrogguGame) GetPlayerScore(i int) int {
	ret := _m.Called(i)
	return ret.Int(0)
}

// GetCardPoints モック
func (_m *MockTrogguGame) GetCardPoints(i int) int {
	ret := _m.Called(i)
	return ret.Int(0)
}

// GetValidPlayIndices モック
func (_m *MockTrogguGame) GetValidPlayIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetGameEndFlag モック
func (_m *MockTrogguGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetWinnerPlayer モック
func (_m *MockTrogguGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHint モック
func (_m *MockTrogguGame) GetHint() *domain.TrogguHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.TrogguHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockTrogguGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

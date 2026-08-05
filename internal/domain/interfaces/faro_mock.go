//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFaroGame はファロゲームのモック。
type MockFaroGame struct {
	mock.Mock
}

func (m *MockFaroGame) Reset() {
	m.Called()
}

func (m *MockFaroGame) NextRound() {
	m.Called()
}

func (m *MockFaroGame) PlayerPlaceBet(rank, amount int, copper bool) error {
	args := m.Called(rank, amount, copper)
	return args.Error(0)
}

func (m *MockFaroGame) PlayerClearBet(rank int) error {
	args := m.Called(rank)
	return args.Error(0)
}

func (m *MockFaroGame) PlayerClearAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockFaroGame) PlayerDealTurn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockFaroGame) PlayerCall(order []int) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockFaroGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFaroGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockFaroGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFaroGame) GetTurnsPlayed() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFaroGame) GetTurnsTotal() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFaroGame) GetRemainingCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFaroGame) GetSoda() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockFaroGame) GetLastTurn() *domain.FaroTurnResult {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.FaroTurnResult)
}

func (m *MockFaroGame) GetCallCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockFaroGame) GetCallOrder() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockFaroGame) GetCallWon() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockFaroGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFaroGame) GetBets() map[int]*domain.FaroBet {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[int]*domain.FaroBet)
}

func (m *MockFaroGame) GetBetRanks() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockFaroGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

// GetRemainingByRank モック
func (m *MockFaroGame) GetRemainingByRank() [domain.FaroMaxRank + 1]int {
	ret := m.Called()
	if v, ok := ret.Get(0).([domain.FaroMaxRank + 1]int); ok {
		return v
	}
	return [domain.FaroMaxRank + 1]int{}
}

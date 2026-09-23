//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBassetGame はバセットゲームのモック。
type MockBassetGame struct {
	mock.Mock
}

func (m *MockBassetGame) Reset() {
	m.Called()
}

func (m *MockBassetGame) NextRound() {
	m.Called()
}

func (m *MockBassetGame) PlayerPlaceBet(rank, amount int) error {
	args := m.Called(rank, amount)
	return args.Error(0)
}

func (m *MockBassetGame) PlayerDealTurn() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBassetGame) PlayerTakeWinnings() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBassetGame) PlayerPressParoli() error { args := m.Called(); return args.Error(0) }

func (m *MockBassetGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBassetGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBassetGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBassetGame) GetTurnsPlayed() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBassetGame) GetTurnsTotal() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBassetGame) GetRemainingCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBassetGame) GetLastTurn() *domain.BassetTurnResult {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.BassetTurnResult)
}

func (m *MockBassetGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBassetGame) GetBet() (*domain.BassetBet, int) {
	args := m.Called()
	var bet *domain.BassetBet
	if args.Get(0) != nil {
		bet = args.Get(0).(*domain.BassetBet)
	}
	return bet, args.Int(1)
}

func (m *MockBassetGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

// GetRemainingByRank モック
func (m *MockBassetGame) GetRemainingByRank() [domain.BassetMaxRank + 1]int {
	ret := m.Called()
	if v, ok := ret.Get(0).([domain.BassetMaxRank + 1]int); ok {
		return v
	}
	return [domain.BassetMaxRank + 1]int{}
}

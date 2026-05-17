//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFourCardPokerGame is the testify mock for FourCardPokerGame.
type MockFourCardPokerGame struct {
	mock.Mock
}

func (m *MockFourCardPokerGame) Reset() {
	m.Called()
}

func (m *MockFourCardPokerGame) Bet(ante, acesUp int) error {
	args := m.Called(ante, acesUp)
	return args.Error(0)
}

func (m *MockFourCardPokerGame) Play(multiplier int) error {
	args := m.Called(multiplier)
	return args.Error(0)
}

func (m *MockFourCardPokerGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockFourCardPokerGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockFourCardPokerGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockFourCardPokerGame) GetPlayerBest() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockFourCardPokerGame) GetDealerBest() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockFourCardPokerGame) GetDealerUpCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockFourCardPokerGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockFourCardPokerGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetAcesUpBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetPlayMultiplier() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockFourCardPokerGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetAnteBonusPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetAcesUpPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockFourCardPokerGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

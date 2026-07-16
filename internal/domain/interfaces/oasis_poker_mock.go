//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOasisPokerGame オアシスポーカーゲームモック
type MockOasisPokerGame struct {
	mock.Mock
}

func (m *MockOasisPokerGame) Reset() {
	m.Called()
}

func (m *MockOasisPokerGame) Bet(ante, jackpot int) error {
	args := m.Called(ante, jackpot)
	return args.Error(0)
}

func (m *MockOasisPokerGame) Exchange(indices []int) error {
	args := m.Called(indices)
	return args.Error(0)
}

func (m *MockOasisPokerGame) Stand() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOasisPokerGame) Play() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOasisPokerGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOasisPokerGame) RecommendPlay() bool {
	return m.Called().Bool(0)
}
func (m *MockOasisPokerGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockOasisPokerGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockOasisPokerGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOasisPokerGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetJackpotBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetExchangeCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetExchangeFee() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockOasisPokerGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetJackpotPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetDealerQualified() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOasisPokerGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOasisPokerGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

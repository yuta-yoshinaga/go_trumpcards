//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRussianPokerGame ロシアンポーカーゲームモック
type MockRussianPokerGame struct {
	mock.Mock
}

func (m *MockRussianPokerGame) Reset() {
	m.Called()
}

func (m *MockRussianPokerGame) Bet(ante int) error {
	args := m.Called(ante)
	return args.Error(0)
}

func (m *MockRussianPokerGame) Exchange(indices []int) error {
	args := m.Called(indices)
	return args.Error(0)
}

func (m *MockRussianPokerGame) Buy6th() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRussianPokerGame) Select(discardIndex int) error {
	args := m.Called(discardIndex)
	return args.Error(0)
}

func (m *MockRussianPokerGame) Play() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRussianPokerGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRussianPokerGame) ForceExchange() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRussianPokerGame) Decline() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRussianPokerGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockRussianPokerGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockRussianPokerGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRussianPokerGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetExchangeCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetExchangeFee() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetBought6th() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRussianPokerGame) GetBuy6thFee() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetForceExchanged() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRussianPokerGame) GetForceExchangeFee() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockRussianPokerGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetDealerQualified() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRussianPokerGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRussianPokerGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

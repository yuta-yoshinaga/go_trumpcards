//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPaiGowGame パイガオポーカーゲームモック
type MockPaiGowGame struct {
	mock.Mock
}

func (m *MockPaiGowGame) Reset() {
	m.Called()
}

func (m *MockPaiGowGame) Bet(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockPaiGowGame) SetHands(lowIdx0, lowIdx1 int) error {
	args := m.Called(lowIdx0, lowIdx1)
	return args.Error(0)
}

func (m *MockPaiGowGame) AutoSetHands() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPaiGowGame) GetHint() *domain.PaiGowHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.PaiGowHint)
}

func (m *MockPaiGowGame) GetPlayerCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockPaiGowGame) GetDealerCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockPaiGowGame) GetPlayerHighHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockPaiGowGame) GetPlayerLowHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockPaiGowGame) GetDealerHighHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockPaiGowGame) GetDealerLowHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockPaiGowGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPaiGowGame) GetBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockPaiGowGame) GetHighHandResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockPaiGowGame) GetLowHandResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockPaiGowGame) GetPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetCommission() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetPlayerHighRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetPlayerLowRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetDealerHighRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetDealerLowRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPaiGowGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

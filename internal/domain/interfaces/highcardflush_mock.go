//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHighCardFlushGame ハイカードフラッシュゲームモック
type MockHighCardFlushGame struct {
	mock.Mock
}

func (m *MockHighCardFlushGame) Reset() {
	m.Called()
}

func (m *MockHighCardFlushGame) Bet(ante, flushBonus, straightFlush int) error {
	args := m.Called(ante, flushBonus, straightFlush)
	return args.Error(0)
}

func (m *MockHighCardFlushGame) Raise(multiplier int) error {
	args := m.Called(multiplier)
	return args.Error(0)
}

func (m *MockHighCardFlushGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockHighCardFlushGame) MaxRaiseMultiplier() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockHighCardFlushGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockHighCardFlushGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockHighCardFlushGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetFlushBonusBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetStraightFlushBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetRaiseBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockHighCardFlushGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetRaisePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetFlushBonusPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetStraightFlushPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetDealerQualified() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockHighCardFlushGame) GetPlayerFlushLen() int {
	args := m.Called()
	return args.Int(0)
}

// GetPlayerFlushSuit モック
func (m *MockHighCardFlushGame) GetPlayerFlushSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetDealerFlushLen() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetPlayerStraightFlushLen() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockHighCardFlushGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

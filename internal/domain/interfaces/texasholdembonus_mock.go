//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTexasHoldemBonusGame テキサスホールデムボーナスポーカーゲームモック
type MockTexasHoldemBonusGame struct {
	mock.Mock
}

func (m *MockTexasHoldemBonusGame) Reset() {
	m.Called()
}

func (m *MockTexasHoldemBonusGame) Bet(ante, bonus int) error {
	args := m.Called(ante, bonus)
	return args.Error(0)
}

func (m *MockTexasHoldemBonusGame) Play() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTexasHoldemBonusGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTexasHoldemBonusGame) Check() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTexasHoldemBonusGame) Raise() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTexasHoldemBonusGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockTexasHoldemBonusGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockTexasHoldemBonusGame) GetCommunity() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockTexasHoldemBonusGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTexasHoldemBonusGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetNextBetCost() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetBonusBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetFlopBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetTurnBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetRiverBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetTotalPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockTexasHoldemBonusGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetBonusPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetPlayerBest() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockTexasHoldemBonusGame) GetDealerBest() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockTexasHoldemBonusGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTexasHoldemBonusGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

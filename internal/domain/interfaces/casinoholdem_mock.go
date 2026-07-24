//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCasinoHoldemGame カジノホールデムゲームモック
type MockCasinoHoldemGame struct {
	mock.Mock
}

func (m *MockCasinoHoldemGame) Reset() {
	m.Called()
}

func (m *MockCasinoHoldemGame) Bet(ante, bonus int) error {
	args := m.Called(ante, bonus)
	return args.Error(0)
}

func (m *MockCasinoHoldemGame) Call() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCasinoHoldemGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCasinoHoldemGame) RecommendCall() bool {
	return m.Called().Bool(0)
}
func (m *MockCasinoHoldemGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCasinoHoldemGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCasinoHoldemGame) GetCommunity() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCasinoHoldemGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCasinoHoldemGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetBonusBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetCallBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockCasinoHoldemGame) GetDealerQualify() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCasinoHoldemGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetCallPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetBonusPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetPlayerBest() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCasinoHoldemGame) GetDealerBest() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCasinoHoldemGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoHoldemGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

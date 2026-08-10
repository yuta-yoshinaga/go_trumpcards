//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockUltimateTexasHoldemGame アルティメット・テキサスホールデムゲームモック
type MockUltimateTexasHoldemGame struct {
	mock.Mock
}

func (m *MockUltimateTexasHoldemGame) Reset() {
	m.Called()
}

func (m *MockUltimateTexasHoldemGame) Bet(ante, trips int) error {
	args := m.Called(ante, trips)
	return args.Error(0)
}

func (m *MockUltimateTexasHoldemGame) Play(multiplier int) error {
	args := m.Called(multiplier)
	return args.Error(0)
}

func (m *MockUltimateTexasHoldemGame) Check() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUltimateTexasHoldemGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUltimateTexasHoldemGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockUltimateTexasHoldemGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockUltimateTexasHoldemGame) GetCommunity() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockUltimateTexasHoldemGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockUltimateTexasHoldemGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetBlindBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetTripsBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetFolded() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockUltimateTexasHoldemGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockUltimateTexasHoldemGame) GetDealerQualified() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockUltimateTexasHoldemGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetBlindPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetTripsPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) RecommendPlay() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUltimateTexasHoldemGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetPlayerBest() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockUltimateTexasHoldemGame) GetDealerBest() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockUltimateTexasHoldemGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockUltimateTexasHoldemGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

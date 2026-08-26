//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockThreeCardRummyGame スリーカード・ラミーゲームモック
type MockThreeCardRummyGame struct {
	mock.Mock
}

func (m *MockThreeCardRummyGame) Reset() {
	m.Called()
}

func (m *MockThreeCardRummyGame) Bet(ante, lowBonus int) error {
	args := m.Called(ante, lowBonus)
	return args.Error(0)
}

// Rebet モック
func (m *MockThreeCardRummyGame) Rebet() error { return m.Called().Error(0) }

func (m *MockThreeCardRummyGame) Play() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockThreeCardRummyGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockThreeCardRummyGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockThreeCardRummyGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockThreeCardRummyGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockThreeCardRummyGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetLowBonusBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockThreeCardRummyGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetAnteBonusPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetLowBonusPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetDealerQualified() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockThreeCardRummyGame) GetPlayerScore() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetDealerScore() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardRummyGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

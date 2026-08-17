//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockThreeCardGame スリーカードポーカーゲームモック
type MockThreeCardGame struct {
	mock.Mock
}

func (m *MockThreeCardGame) Reset() {
	m.Called()
}

func (m *MockThreeCardGame) Bet(ante, pairPlus int) error {
	args := m.Called(ante, pairPlus)
	return args.Error(0)
}

// Rebet モック
func (m *MockThreeCardGame) Rebet() error { return m.Called().Error(0) }

func (m *MockThreeCardGame) Play() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockThreeCardGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockThreeCardGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockThreeCardGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockThreeCardGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockThreeCardGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetPairPlusBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockThreeCardGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetAnteBonusPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetPairPlusPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetDealerQualified() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockThreeCardGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockThreeCardGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

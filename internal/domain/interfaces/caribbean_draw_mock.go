//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCaribbeanDrawGame カリビアン・ドロー・ポーカーゲームモック
type MockCaribbeanDrawGame struct {
	mock.Mock
}

func (m *MockCaribbeanDrawGame) Reset() {
	m.Called()
}

func (m *MockCaribbeanDrawGame) Bet(ante, jackpot int) error {
	args := m.Called(ante, jackpot)
	return args.Error(0)
}

func (m *MockCaribbeanDrawGame) Draw(indices []int) error {
	args := m.Called(indices)
	return args.Error(0)
}

func (m *MockCaribbeanDrawGame) Play() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCaribbeanDrawGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCaribbeanDrawGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCaribbeanDrawGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCaribbeanDrawGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCaribbeanDrawGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetJackpotBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockCaribbeanDrawGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetDrawCost() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetJackpotPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetDealerQualified() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCaribbeanDrawGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanDrawGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

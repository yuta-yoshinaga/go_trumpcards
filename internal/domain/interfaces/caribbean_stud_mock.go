//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCaribbeanStudGame カリビアンスタッドポーカーゲームモック
type MockCaribbeanStudGame struct {
	mock.Mock
}

func (m *MockCaribbeanStudGame) Reset() {
	m.Called()
}

func (m *MockCaribbeanStudGame) Bet(ante, jackpot int) error {
	args := m.Called(ante, jackpot)
	return args.Error(0)
}

func (m *MockCaribbeanStudGame) Play() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCaribbeanStudGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCaribbeanStudGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCaribbeanStudGame) GetDealerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCaribbeanStudGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCaribbeanStudGame) GetAnteBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetJackpotBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetPlayBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockCaribbeanStudGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetPlayPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetJackpotPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetDealerQualified() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCaribbeanStudGame) GetPlayerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetDealerHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCaribbeanStudGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

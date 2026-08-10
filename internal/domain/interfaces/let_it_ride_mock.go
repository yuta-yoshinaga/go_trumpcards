//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLetItRideGame レット・イット・ライドゲームモック
type MockLetItRideGame struct {
	mock.Mock
}

func (m *MockLetItRideGame) Reset() {
	m.Called()
}

func (m *MockLetItRideGame) Bet(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockLetItRideGame) Pull() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLetItRideGame) GetPullPreview() *domain.LetItRidePullPreview {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.LetItRidePullPreview)
}

func (m *MockLetItRideGame) LetItRideAction() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLetItRideGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockLetItRideGame) GetCommunityCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockLetItRideGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockLetItRideGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLetItRideGame) GetBetAmount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockLetItRideGame) GetBet1Active() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLetItRideGame) GetBet2Active() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLetItRideGame) GetBet3Active() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLetItRideGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockLetItRideGame) GetHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockLetItRideGame) GetBet1Payout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockLetItRideGame) GetBet2Payout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockLetItRideGame) GetBet3Payout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockLetItRideGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockLetItRideGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockLetItRideGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

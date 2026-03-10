package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBaccaratGame バカラゲームモック
type MockBaccaratGame struct {
	mock.Mock
}

func (m *MockBaccaratGame) Reset() {
	m.Called()
}

func (m *MockBaccaratGame) Bet(amount, betType int) error {
	args := m.Called(amount, betType)
	return args.Error(0)
}

func (m *MockBaccaratGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockBaccaratGame) GetBankerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockBaccaratGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBaccaratGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBaccaratGame) GetBetAmount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBaccaratGame) GetBetType() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBaccaratGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockBaccaratGame) GetPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBaccaratGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBaccaratGame) GetPlayerHandValue() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBaccaratGame) GetBankerHandValue() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBaccaratGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

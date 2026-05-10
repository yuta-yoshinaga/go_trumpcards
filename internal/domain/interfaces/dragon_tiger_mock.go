//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDragonTigerGame ドラゴンタイガーゲームモック
type MockDragonTigerGame struct {
	mock.Mock
}

func (m *MockDragonTigerGame) Reset() {
	m.Called()
}

func (m *MockDragonTigerGame) ClearHistory() {
	m.Called()
}

func (m *MockDragonTigerGame) Bet(amount, betType int) error {
	args := m.Called(amount, betType)
	return args.Error(0)
}

func (m *MockDragonTigerGame) GetDragonCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockDragonTigerGame) GetTigerCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockDragonTigerGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockDragonTigerGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDragonTigerGame) GetBetAmount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockDragonTigerGame) GetBetType() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockDragonTigerGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockDragonTigerGame) GetPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockDragonTigerGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockDragonTigerGame) GetHistory() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockDragonTigerGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

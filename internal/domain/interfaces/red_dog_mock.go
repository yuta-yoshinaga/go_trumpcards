//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRedDogGame レッドドッグゲームモック
type MockRedDogGame struct {
	mock.Mock
}

func (m *MockRedDogGame) Reset() {
	m.Called()
}

func (m *MockRedDogGame) Bet(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockRedDogGame) ResolveInitial() {
	m.Called()
}

func (m *MockRedDogGame) Raise(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockRedDogGame) Stay() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedDogGame) ResolveThird() {
	m.Called()
}

func (m *MockRedDogGame) GetInitialCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockRedDogGame) GetThirdCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockRedDogGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRedDogGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRedDogGame) GetAnte() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRedDogGame) GetRaise() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRedDogGame) GetSpread() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRedDogGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockRedDogGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRedDogGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRedDogGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

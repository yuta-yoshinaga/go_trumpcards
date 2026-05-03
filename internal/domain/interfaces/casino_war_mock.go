//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCasinoWarGame カジノウォーゲームモック
type MockCasinoWarGame struct {
	mock.Mock
}

func (m *MockCasinoWarGame) Reset() {
	m.Called()
}

func (m *MockCasinoWarGame) Bet(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockCasinoWarGame) ResolveInitial() {
	m.Called()
}

func (m *MockCasinoWarGame) Surrender() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCasinoWarGame) GoToWar() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCasinoWarGame) ResolveWar() {
	m.Called()
}

func (m *MockCasinoWarGame) GetPlayerCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockCasinoWarGame) GetDealerCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockCasinoWarGame) GetPlayerWarCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockCasinoWarGame) GetDealerWarCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockCasinoWarGame) GetBurnCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCasinoWarGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoWarGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCasinoWarGame) GetAnte() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoWarGame) GetWarBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoWarGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockCasinoWarGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoWarGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCasinoWarGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOichoKabuGame おいちょかぶゲームモック
type MockOichoKabuGame struct {
	mock.Mock
}

func (m *MockOichoKabuGame) Reset() {
	m.Called()
}

func (m *MockOichoKabuGame) Bet(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockOichoKabuGame) Draw() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOichoKabuGame) Stand() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOichoKabuGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockOichoKabuGame) GetBankerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockOichoKabuGame) GetPlayerRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOichoKabuGame) GetBankerRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOichoKabuGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOichoKabuGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOichoKabuGame) GetBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOichoKabuGame) GetResult() domain.OichoKabuResult {
	args := m.Called()
	return args.Get(0).(domain.OichoKabuResult)
}

func (m *MockOichoKabuGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOichoKabuGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOichoKabuGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

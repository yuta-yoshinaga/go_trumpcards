//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAndarBaharGame アンダーバハールゲームモック
type MockAndarBaharGame struct {
	mock.Mock
}

func (m *MockAndarBaharGame) Reset() {
	m.Called()
}

func (m *MockAndarBaharGame) ClearHistory() {
	m.Called()
}

func (m *MockAndarBaharGame) Bet(amount, target, sideAmount, sideBand int) error {
	args := m.Called(amount, target, sideAmount, sideBand)
	return args.Error(0)
}

func (m *MockAndarBaharGame) GetJoker() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockAndarBaharGame) GetAndarCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockAndarBaharGame) GetBaharCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockAndarBaharGame) GetFirstColumn() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) DealtCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAndarBaharGame) GetBetAmount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetBetTarget() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetSideAmount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetSideBand() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetWinner() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockAndarBaharGame) GetPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAndarBaharGame) GetHistory() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockAndarBaharGame) GetHint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAndarBaharGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

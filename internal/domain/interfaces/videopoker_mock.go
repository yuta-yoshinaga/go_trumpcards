//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockVideoPokerGame ビデオポーカーゲームモック
type MockVideoPokerGame struct {
	mock.Mock
}

func (m *MockVideoPokerGame) Reset() {
	m.Called()
}

func (m *MockVideoPokerGame) Bet(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockVideoPokerGame) Hold(indices []int) error {
	args := m.Called(indices)
	return args.Error(0)
}

func (m *MockVideoPokerGame) GetHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockVideoPokerGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockVideoPokerGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockVideoPokerGame) GetBetAmount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockVideoPokerGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockVideoPokerGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockVideoPokerGame) GetPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockVideoPokerGame) GetHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockVideoPokerGame) GetHandName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockVideoPokerGame) GetHandKey() string {
	args := m.Called()
	return args.String(0)
}

// GetCurrentHandKey モック
func (m *MockVideoPokerGame) GetCurrentHandKey() string { return m.Called().String(0) }

func (m *MockVideoPokerGame) GetHeldIndices() [domain.VideoPokerHandSize]bool {
	args := m.Called()
	return args.Get(0).([domain.VideoPokerHandSize]bool)
}

func (m *MockVideoPokerGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

func (m *MockVideoPokerGame) GetVariantName() string {
	args := m.Called()
	return args.String(0)
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMississippiStudGame ミシシッピ・スタッドゲームモック
type MockMississippiStudGame struct {
	mock.Mock
}

func (m *MockMississippiStudGame) Reset() {
	m.Called()
}

func (m *MockMississippiStudGame) Bet(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockMississippiStudGame) Play(multiplier int) error {
	args := m.Called(multiplier)
	return args.Error(0)
}

func (m *MockMississippiStudGame) Fold() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMississippiStudGame) GetPlayerHand() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockMississippiStudGame) GetCommunityCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockMississippiStudGame) GetCommunityRevealed() [domain.MississippiStudCommunityCnt]bool {
	args := m.Called()
	return args.Get(0).([domain.MississippiStudCommunityCnt]bool)
}

func (m *MockMississippiStudGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMississippiStudGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMississippiStudGame) GetAnteAmount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMississippiStudGame) GetStreetMultipliers() [domain.MississippiStudStreetCnt]int {
	args := m.Called()
	return args.Get(0).([domain.MississippiStudStreetCnt]int)
}

func (m *MockMississippiStudGame) GetFolded() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockMississippiStudGame) GetTotalBet() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMississippiStudGame) GetResult() domain.GameResult {
	args := m.Called()
	return args.Get(0).(domain.GameResult)
}

func (m *MockMississippiStudGame) GetHandRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMississippiStudGame) GetCurrentMadeHand() *domain.MississippiStudMadeHand {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.MississippiStudMadeHand)
}

func (m *MockMississippiStudGame) RecommendBet() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockMississippiStudGame) GetPayoutMultiplier() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMississippiStudGame) GetAntePayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMississippiStudGame) GetStreetPayouts() [domain.MississippiStudStreetCnt]int {
	args := m.Called()
	return args.Get(0).([domain.MississippiStudStreetCnt]int)
}

func (m *MockMississippiStudGame) GetTotalPayout() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMississippiStudGame) GetChips() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockMississippiStudGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

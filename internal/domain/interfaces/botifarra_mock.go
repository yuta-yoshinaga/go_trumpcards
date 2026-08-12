//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBotifarraGame ボティファラゲームモック
type MockBotifarraGame struct {
	mock.Mock
}

func (m *MockBotifarraGame) Reset() {
	m.Called()
}

func (m *MockBotifarraGame) Declare(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockBotifarraGame) Delegate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBotifarraGame) Double() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBotifarraGame) PassDouble() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBotifarraGame) PlayCard(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBotifarraGame) NextRound() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBotifarraGame) GiveUp() {
	m.Called()
}

func (m *MockBotifarraGame) CpuPlay() {
	m.Called()
}

func (m *MockBotifarraGame) GetConfig() domain.BotifarraConfig {
	args := m.Called()
	return args.Get(0).(domain.BotifarraConfig)
}

func (m *MockBotifarraGame) SetConfig(cfg domain.BotifarraConfig) {
	m.Called(cfg)
}

func (m *MockBotifarraGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBotifarraGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBotifarraGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockBotifarraGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetDeclarerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetMultiplier() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetCurrentTurn() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockBotifarraGame) GetLastTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockBotifarraGame) GetLastTrickWinner() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetTrickCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetRoundPoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockBotifarraGame) GetScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockBotifarraGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBotifarraGame) GetPlayer(i int) *domain.BotifarraPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.BotifarraPlayer)
}

func (m *MockBotifarraGame) GetHint() *domain.BotifarraHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.BotifarraHint)
}

func (m *MockBotifarraGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

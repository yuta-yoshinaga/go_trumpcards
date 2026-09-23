//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBatakGame Batak ゲームモック
type MockBatakGame struct {
	mock.Mock
}

func (m *MockBatakGame) Reset() { m.Called() }

func (m *MockBatakGame) NextRound() { m.Called() }

func (m *MockBatakGame) PlayerBid(bid int) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockBatakGame) CpuBid() { m.Called() }

func (m *MockBatakGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBatakGame) CpuPlay() { m.Called() }

func (m *MockBatakGame) ResolveTrick() { m.Called() }

func (m *MockBatakGame) NextTrick() { m.Called() }

func (m *MockBatakGame) ScoreRound() { m.Called() }

func (m *MockBatakGame) GetConfig() domain.BatakConfig {
	args := m.Called()
	return args.Get(0).(domain.BatakConfig)
}

func (m *MockBatakGame) SetConfig(cfg domain.BatakConfig) { m.Called(cfg) }

func (m *MockBatakGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBatakGame) GetPhase() domain.BatakPhase {
	args := m.Called()
	return args.Get(0).(domain.BatakPhase)
}

func (m *MockBatakGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBatakGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBatakGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockBatakGame) GetSpadesBroken() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBatakGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetDeclarerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetHighBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetBidStartIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) MinLegalBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBatakGame) GetPlayer(i int) *domain.BatakPlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.BatakPlayer)
}

// GetHint モック
func (m *MockBatakGame) GetHint() *domain.BatakHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.BatakHint); ok {
		return val
	}
	return nil
}

func (m *MockBatakGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}

// GetValidPlayIndices モック
func (m *MockBatakGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if val, ok := args.Get(0).([]int); ok {
		return val
	}
	return nil
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCallBreakGame Call Break ゲームモック
type MockCallBreakGame struct {
	mock.Mock
}

func (m *MockCallBreakGame) Reset() { m.Called() }

func (m *MockCallBreakGame) NextRound() { m.Called() }

func (m *MockCallBreakGame) PlayerBid(bid int) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockCallBreakGame) CpuBid() { m.Called() }

func (m *MockCallBreakGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockCallBreakGame) CpuPlay() { m.Called() }

func (m *MockCallBreakGame) ResolveTrick() { m.Called() }

func (m *MockCallBreakGame) NextTrick() { m.Called() }

func (m *MockCallBreakGame) ScoreRound() { m.Called() }

func (m *MockCallBreakGame) GetConfig() domain.CallBreakConfig {
	args := m.Called()
	return args.Get(0).(domain.CallBreakConfig)
}

func (m *MockCallBreakGame) SetConfig(cfg domain.CallBreakConfig) { m.Called(cfg) }

func (m *MockCallBreakGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCallBreakGame) GetPhase() domain.CallBreakPhase {
	args := m.Called()
	return args.Get(0).(domain.CallBreakPhase)
}

func (m *MockCallBreakGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCallBreakGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCallBreakGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCallBreakGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCallBreakGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCallBreakGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockCallBreakGame) GetSpadesBroken() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCallBreakGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCallBreakGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCallBreakGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCallBreakGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCallBreakGame) GetPlayer(i int) *domain.CallBreakPlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.CallBreakPlayer)
}

// GetHint モック
func (m *MockCallBreakGame) GetHint() *domain.CallBreakHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.CallBreakHint); ok {
		return val
	}
	return nil
}

func (m *MockCallBreakGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}

// GetValidPlayIndices モック
func (m *MockCallBreakGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if val, ok := args.Get(0).([]int); ok {
		return val
	}
	return nil
}

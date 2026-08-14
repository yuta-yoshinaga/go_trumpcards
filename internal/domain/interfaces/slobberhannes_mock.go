//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSlobberhannesGame スロバーハンネスゲームモック
type MockSlobberhannesGame struct {
	mock.Mock
}

func (m *MockSlobberhannesGame) Reset() { m.Called() }

func (m *MockSlobberhannesGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockSlobberhannesGame) CpuPlay()   { m.Called() }
func (m *MockSlobberhannesGame) NextRound() { m.Called() }
func (m *MockSlobberhannesGame) GiveUp()    { m.Called() }

func (m *MockSlobberhannesGame) GetConfig() domain.SlobberhannesConfig {
	return m.Called().Get(0).(domain.SlobberhannesConfig)
}

func (m *MockSlobberhannesGame) SetConfig(cfg domain.SlobberhannesConfig) { m.Called(cfg) }

func (m *MockSlobberhannesGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockSlobberhannesGame) GetPhase() domain.SlobberhannesPhase {
	return m.Called().Get(0).(domain.SlobberhannesPhase)
}

func (m *MockSlobberhannesGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockSlobberhannesGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockSlobberhannesGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockSlobberhannesGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockSlobberhannesGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockSlobberhannesGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockSlobberhannesGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockSlobberhannesGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockSlobberhannesGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockSlobberhannesGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockSlobberhannesGame) GetPlayer(i int) *domain.SlobberhannesPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.SlobberhannesPlayer)
}

func (m *MockSlobberhannesGame) GetHint() *domain.SlobberhannesHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.SlobberhannesHint)
}

func (m *MockSlobberhannesGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockReversisGame レヴェルシゲームモック
type MockReversisGame struct {
	mock.Mock
}

func (m *MockReversisGame) Reset()     { m.Called() }
func (m *MockReversisGame) CpuPlay()   { m.Called() }
func (m *MockReversisGame) NextRound() { m.Called() }
func (m *MockReversisGame) GiveUp()    { m.Called() }

func (m *MockReversisGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockReversisGame) GetConfig() domain.ReversisConfig {
	return m.Called().Get(0).(domain.ReversisConfig)
}

func (m *MockReversisGame) SetConfig(cfg domain.ReversisConfig) { m.Called(cfg) }

func (m *MockReversisGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockReversisGame) GetPhase() domain.ReversisPhase {
	return m.Called().Get(0).(domain.ReversisPhase)
}

func (m *MockReversisGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockReversisGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockReversisGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockReversisGame) GetPool() int             { return m.Called().Int(0) }
func (m *MockReversisGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockReversisGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockReversisGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockReversisGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockReversisGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockReversisGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockReversisGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockReversisGame) GetPlayer(i int) *domain.ReversisPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ReversisPlayer)
}

func (m *MockReversisGame) GetHint() *domain.ReversisHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ReversisHint)
}

func (m *MockReversisGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

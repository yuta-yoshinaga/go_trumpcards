//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRamsGame ラムスゲームモック
type MockRamsGame struct {
	mock.Mock
}

func (m *MockRamsGame) Reset()     { m.Called() }
func (m *MockRamsGame) CpuPlay()   { m.Called() }
func (m *MockRamsGame) NextRound() { m.Called() }
func (m *MockRamsGame) GiveUp()    { m.Called() }

func (m *MockRamsGame) Decide(play bool) error { return m.Called(play).Error(0) }

func (m *MockRamsGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockRamsGame) GetConfig() domain.RamsConfig {
	return m.Called().Get(0).(domain.RamsConfig)
}

func (m *MockRamsGame) SetConfig(cfg domain.RamsConfig) { m.Called(cfg) }

func (m *MockRamsGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockRamsGame) GetPhase() domain.RamsPhase {
	return m.Called().Get(0).(domain.RamsPhase)
}

func (m *MockRamsGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockRamsGame) IsDecidePhase() bool      { return m.Called().Bool(0) }
func (m *MockRamsGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockRamsGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockRamsGame) GetPot() int              { return m.Called().Int(0) }
func (m *MockRamsGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockRamsGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockRamsGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockRamsGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockRamsGame) GetActiveCount() int      { return m.Called().Int(0) }
func (m *MockRamsGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockRamsGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockRamsGame) GetUpCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockRamsGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockRamsGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockRamsGame) GetPlayer(i int) *domain.RamsPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.RamsPlayer)
}

func (m *MockRamsGame) GetHint() *domain.RamsHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.RamsHint)
}

func (m *MockRamsGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

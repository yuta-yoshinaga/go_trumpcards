//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockJulepeGame フレペゲームモック
type MockJulepeGame struct {
	mock.Mock
}

func (m *MockJulepeGame) Reset()     { m.Called() }
func (m *MockJulepeGame) CpuPlay()   { m.Called() }
func (m *MockJulepeGame) NextRound() { m.Called() }
func (m *MockJulepeGame) GiveUp()    { m.Called() }

func (m *MockJulepeGame) Decide(play bool) error { return m.Called(play).Error(0) }

func (m *MockJulepeGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockJulepeGame) GetConfig() domain.JulepeConfig {
	return m.Called().Get(0).(domain.JulepeConfig)
}

func (m *MockJulepeGame) SetConfig(cfg domain.JulepeConfig) { m.Called(cfg) }

func (m *MockJulepeGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockJulepeGame) GetPhase() domain.JulepePhase {
	return m.Called().Get(0).(domain.JulepePhase)
}

func (m *MockJulepeGame) IsHumanTurn() bool      { return m.Called().Bool(0) }
func (m *MockJulepeGame) IsDecidePhase() bool    { return m.Called().Bool(0) }
func (m *MockJulepeGame) GetRoundNumber() int    { return m.Called().Int(0) }
func (m *MockJulepeGame) GetTrickNumber() int    { return m.Called().Int(0) }
func (m *MockJulepeGame) GetPot() int            { return m.Called().Int(0) }
func (m *MockJulepeGame) GetRequiredTricks() int { return m.Called().Int(0) }

func (m *MockJulepeGame) GetBeast() []bool {
	if v := m.Called().Get(0); v != nil {
		return v.([]bool)
	}
	return nil
}
func (m *MockJulepeGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockJulepeGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockJulepeGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockJulepeGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockJulepeGame) GetActiveCount() int      { return m.Called().Int(0) }
func (m *MockJulepeGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockJulepeGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockJulepeGame) GetUpCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockJulepeGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockJulepeGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockJulepeGame) GetPlayer(i int) *domain.JulepePlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.JulepePlayer)
}

func (m *MockJulepeGame) GetHint() *domain.JulepeHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.JulepeHint)
}

func (m *MockJulepeGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

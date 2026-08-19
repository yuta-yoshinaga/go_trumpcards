//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTeenDoPaanchGame 3-2-5 ゲームモック
type MockTeenDoPaanchGame struct {
	mock.Mock
}

func (m *MockTeenDoPaanchGame) Reset()           { m.Called() }
func (m *MockTeenDoPaanchGame) CpuDeclareTrump() { m.Called() }
func (m *MockTeenDoPaanchGame) CpuPlay()         { m.Called() }
func (m *MockTeenDoPaanchGame) NextRound()       { m.Called() }
func (m *MockTeenDoPaanchGame) GiveUp()          { m.Called() }

func (m *MockTeenDoPaanchGame) PlayerDeclareTrump(suit int) error {
	return m.Called(suit).Error(0)
}

func (m *MockTeenDoPaanchGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockTeenDoPaanchGame) GetConfig() domain.TeenDoPaanchConfig {
	return m.Called().Get(0).(domain.TeenDoPaanchConfig)
}

func (m *MockTeenDoPaanchGame) SetConfig(cfg domain.TeenDoPaanchConfig) { m.Called(cfg) }

func (m *MockTeenDoPaanchGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockTeenDoPaanchGame) GetPhase() domain.TeenDoPaanchPhase {
	return m.Called().Get(0).(domain.TeenDoPaanchPhase)
}

func (m *MockTeenDoPaanchGame) IsHumanTurn() bool      { return m.Called().Bool(0) }
func (m *MockTeenDoPaanchGame) IsHumanTrumpTurn() bool { return m.Called().Bool(0) }
func (m *MockTeenDoPaanchGame) GetRoundNumber() int    { return m.Called().Int(0) }
func (m *MockTeenDoPaanchGame) GetTrickNumber() int    { return m.Called().Int(0) }
func (m *MockTeenDoPaanchGame) GetTrumpSuit() int      { return m.Called().Int(0) }
func (m *MockTeenDoPaanchGame) GetFivePlayerIdx() int  { return m.Called().Int(0) }
func (m *MockTeenDoPaanchGame) GetLastExchange() int   { return m.Called().Int(0) }

// GetLastExchangePairs モック
func (m *MockTeenDoPaanchGame) GetLastExchangePairs() []domain.TeenDoPaanchExchange {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]domain.TeenDoPaanchExchange)
}
func (m *MockTeenDoPaanchGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockTeenDoPaanchGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockTeenDoPaanchGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockTeenDoPaanchGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockTeenDoPaanchGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockTeenDoPaanchGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockTeenDoPaanchGame) GetPlayer(i int) *domain.TeenDoPaanchPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.TeenDoPaanchPlayer)
	}
	return nil
}

func (m *MockTeenDoPaanchGame) GetHint() *domain.TeenDoPaanchHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.TeenDoPaanchHint)
	}
	return nil
}

func (m *MockTeenDoPaanchGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

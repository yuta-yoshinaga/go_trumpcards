//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHokmGame ホクムゲームモック
type MockHokmGame struct {
	mock.Mock
}

func (m *MockHokmGame) Reset()           { m.Called() }
func (m *MockHokmGame) CpuDeclareTrump() { m.Called() }
func (m *MockHokmGame) CpuPlay()         { m.Called() }
func (m *MockHokmGame) NextHand()        { m.Called() }
func (m *MockHokmGame) GiveUp()          { m.Called() }

func (m *MockHokmGame) PlayerDeclareTrump(suit int) error { return m.Called(suit).Error(0) }

func (m *MockHokmGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockHokmGame) GetConfig() domain.HokmConfig {
	return m.Called().Get(0).(domain.HokmConfig)
}

func (m *MockHokmGame) SetConfig(cfg domain.HokmConfig) { m.Called(cfg) }

func (m *MockHokmGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockHokmGame) GetPhase() domain.HokmPhase {
	return m.Called().Get(0).(domain.HokmPhase)
}

func (m *MockHokmGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockHokmGame) IsHumanTrumpTurn() bool   { return m.Called().Bool(0) }
func (m *MockHokmGame) GetHandNumber() int       { return m.Called().Int(0) }
func (m *MockHokmGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockHokmGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockHokmGame) GetHakemIdx() int         { return m.Called().Int(0) }
func (m *MockHokmGame) GetLastHandKot() bool     { return m.Called().Bool(0) }
func (m *MockHokmGame) GetLastHandWinner() int   { return m.Called().Int(0) }
func (m *MockHokmGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockHokmGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockHokmGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockHokmGame) GetWinnerTeam() int       { return m.Called().Int(0) }

func (m *MockHokmGame) GetScore(team int) int   { return m.Called(team).Int(0) }
func (m *MockHokmGame) TeamTricks(team int) int { return m.Called(team).Int(0) }

func (m *MockHokmGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockHokmGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockHokmGame) GetPlayer(i int) *domain.HokmPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.HokmPlayer)
}

func (m *MockHokmGame) GetHint() *domain.HokmHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.HokmHint)
}

func (m *MockHokmGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

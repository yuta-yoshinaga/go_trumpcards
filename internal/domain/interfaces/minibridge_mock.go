//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMinibridgeGame ミニブリッジゲームモック
type MockMinibridgeGame struct {
	mock.Mock
}

func (m *MockMinibridgeGame) Reset()             { m.Called() }
func (m *MockMinibridgeGame) CpuSelectContract() { m.Called() }
func (m *MockMinibridgeGame) CpuPlay()           { m.Called() }
func (m *MockMinibridgeGame) NextRound()         { m.Called() }
func (m *MockMinibridgeGame) GiveUp()            { m.Called() }

func (m *MockMinibridgeGame) PlayerSelectContract(level, suit int) error {
	return m.Called(level, suit).Error(0)
}

func (m *MockMinibridgeGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockMinibridgeGame) GetConfig() domain.MinibridgeConfig {
	return m.Called().Get(0).(domain.MinibridgeConfig)
}

func (m *MockMinibridgeGame) SetConfig(cfg domain.MinibridgeConfig) { m.Called(cfg) }

func (m *MockMinibridgeGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockMinibridgeGame) GetPhase() domain.MinibridgePhase {
	return m.Called().Get(0).(domain.MinibridgePhase)
}

func (m *MockMinibridgeGame) IsHumanTurn() bool         { return m.Called().Bool(0) }
func (m *MockMinibridgeGame) IsHumanContractTurn() bool { return m.Called().Bool(0) }
func (m *MockMinibridgeGame) GetRoundNumber() int       { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetTrickNumber() int       { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetContractLevel() int     { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetContractSuit() int      { return m.Called().Int(0) }
func (m *MockMinibridgeGame) RequiredTricks() int       { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetDeclarerIdx() int       { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetDummyIdx() int          { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetLastMade() bool         { return m.Called().Bool(0) }
func (m *MockMinibridgeGame) GetLastTricks() int        { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetCurrentPlayerIdx() int  { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetLeadPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetDealerIdx() int         { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetPlayerCnt() int         { return m.Called().Int(0) }
func (m *MockMinibridgeGame) GetWinnerTeam() int        { return m.Called().Int(0) }

func (m *MockMinibridgeGame) GetTeamScore(team int) int { return m.Called(team).Int(0) }

func (m *MockMinibridgeGame) GetDummyHand() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockMinibridgeGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockMinibridgeGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockMinibridgeGame) GetPlayer(i int) *domain.MinibridgePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.MinibridgePlayer)
	}
	return nil
}

func (m *MockMinibridgeGame) GetHint() *domain.MinibridgeHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.MinibridgeHint)
	}
	return nil
}

func (m *MockMinibridgeGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

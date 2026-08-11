//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBalootGame バルートゲームモック
type MockBalootGame struct {
	mock.Mock
}

func (m *MockBalootGame) Reset()      { m.Called() }
func (m *MockBalootGame) CpuDeclare() { m.Called() }
func (m *MockBalootGame) CpuPlay()    { m.Called() }
func (m *MockBalootGame) NextRound()  { m.Called() }
func (m *MockBalootGame) GiveUp()     { m.Called() }

func (m *MockBalootGame) DeclareSun() error           { return m.Called().Error(0) }
func (m *MockBalootGame) DeclareHokom(suit int) error { return m.Called(suit).Error(0) }
func (m *MockBalootGame) PassDeclaration() error      { return m.Called().Error(0) }
func (m *MockBalootGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockBalootGame) GetConfig() domain.BalootConfig {
	return m.Called().Get(0).(domain.BalootConfig)
}

func (m *MockBalootGame) SetConfig(cfg domain.BalootConfig) { m.Called(cfg) }

func (m *MockBalootGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockBalootGame) GetPhase() domain.BalootPhase {
	return m.Called().Get(0).(domain.BalootPhase)
}

func (m *MockBalootGame) GetMode() domain.BalootMode {
	return m.Called().Get(0).(domain.BalootMode)
}

func (m *MockBalootGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockBalootGame) IsHumanDeclareTurn() bool { return m.Called().Bool(0) }
func (m *MockBalootGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockBalootGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockBalootGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockBalootGame) GetDeclarerIdx() int      { return m.Called().Int(0) }
func (m *MockBalootGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockBalootGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockBalootGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockBalootGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockBalootGame) GetWinnerTeam() int       { return m.Called().Int(0) }

func (m *MockBalootGame) GetScore(team int) int       { return m.Called(team).Int(0) }
func (m *MockBalootGame) GetRoundPoints(team int) int { return m.Called(team).Int(0) }

func (m *MockBalootGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockBalootGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockBalootGame) GetPlayer(i int) *domain.BalootPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.BalootPlayer)
}

func (m *MockBalootGame) GetHint() *domain.BalootHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.BalootHint)
}

func (m *MockBalootGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

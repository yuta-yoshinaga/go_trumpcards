//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCucumberGame キューカンバーゲームモック
type MockCucumberGame struct {
	mock.Mock
}

func (m *MockCucumberGame) Reset()   { m.Called() }
func (m *MockCucumberGame) CpuPlay() { m.Called() }
func (m *MockCucumberGame) GiveUp()  { m.Called() }

func (m *MockCucumberGame) PlayerPlay(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockCucumberGame) NextRound() error               { return m.Called().Error(0) }

func (m *MockCucumberGame) GetConfig() domain.CucumberConfig {
	return m.Called().Get(0).(domain.CucumberConfig)
}

func (m *MockCucumberGame) SetConfig(cfg domain.CucumberConfig) { m.Called(cfg) }

func (m *MockCucumberGame) GetPhase() domain.CucumberPhase {
	return m.Called().Get(0).(domain.CucumberPhase)
}

func (m *MockCucumberGame) GetGameEndFlag() bool { return m.Called().Bool(0) }
func (m *MockCucumberGame) IsHumanTurn() bool    { return m.Called().Bool(0) }
func (m *MockCucumberGame) HighestInTrick() int  { return m.Called().Int(0) }

func (m *MockCucumberGame) IsForcedLowest(playerIdx int) bool {
	return m.Called(playerIdx).Bool(0)
}

func (m *MockCucumberGame) GetCurrentPlayerIdx() int   { return m.Called().Int(0) }
func (m *MockCucumberGame) GetLeadPlayerIdx() int      { return m.Called().Int(0) }
func (m *MockCucumberGame) GetTrickNumber() int        { return m.Called().Int(0) }
func (m *MockCucumberGame) GetRoundNumber() int        { return m.Called().Int(0) }
func (m *MockCucumberGame) GetLastTrickWinnerIdx() int { return m.Called().Int(0) }
func (m *MockCucumberGame) GetLastPenalty() int        { return m.Called().Int(0) }
func (m *MockCucumberGame) GetPlayerCnt() int          { return m.Called().Int(0) }
func (m *MockCucumberGame) GetWinnerIdx() int          { return m.Called().Int(0) }

func (m *MockCucumberGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockCucumberGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockCucumberGame) GetPlayer(i int) *domain.CucumberPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.CucumberPlayer)
	}
	return nil
}

func (m *MockCucumberGame) GetHint() *domain.CucumberHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.CucumberHint)
	}
	return nil
}

func (m *MockCucumberGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

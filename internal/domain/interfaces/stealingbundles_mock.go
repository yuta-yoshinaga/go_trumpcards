//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockStealingBundlesGame スティーリングバンドルゲームモック
type MockStealingBundlesGame struct {
	mock.Mock
}

func (m *MockStealingBundlesGame) Reset()   { m.Called() }
func (m *MockStealingBundlesGame) CpuPlay() { m.Called() }
func (m *MockStealingBundlesGame) GiveUp()  { m.Called() }

func (m *MockStealingBundlesGame) PlayerTake(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockStealingBundlesGame) PlayerSteal(cardIndex, victimIdx int) error {
	return m.Called(cardIndex, victimIdx).Error(0)
}

func (m *MockStealingBundlesGame) PlayerTrail(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockStealingBundlesGame) GetConfig() domain.StealingBundlesConfig {
	return m.Called().Get(0).(domain.StealingBundlesConfig)
}

func (m *MockStealingBundlesGame) SetConfig(cfg domain.StealingBundlesConfig) { m.Called(cfg) }

func (m *MockStealingBundlesGame) GetPhase() domain.StealingBundlesPhase {
	return m.Called().Get(0).(domain.StealingBundlesPhase)
}

func (m *MockStealingBundlesGame) GetGameEndFlag() bool          { return m.Called().Bool(0) }
func (m *MockStealingBundlesGame) IsHumanTurn() bool             { return m.Called().Bool(0) }
func (m *MockStealingBundlesGame) CanCapture(playerIdx int) bool { return m.Called(playerIdx).Bool(0) }
func (m *MockStealingBundlesGame) GetDeckRemaining() int         { return m.Called().Int(0) }
func (m *MockStealingBundlesGame) GetLastCaptureIdx() int        { return m.Called().Int(0) }
func (m *MockStealingBundlesGame) GetCurrentPlayerIdx() int      { return m.Called().Int(0) }
func (m *MockStealingBundlesGame) GetTurnNumber() int            { return m.Called().Int(0) }
func (m *MockStealingBundlesGame) GetPacksDealt() int            { return m.Called().Int(0) }
func (m *MockStealingBundlesGame) GetPlayerCnt() int             { return m.Called().Int(0) }
func (m *MockStealingBundlesGame) GetWinnerIdx() int             { return m.Called().Int(0) }

func (m *MockStealingBundlesGame) GetTableMatches(playerIdx, cardIndex int) []int {
	args := m.Called(playerIdx, cardIndex)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockStealingBundlesGame) GetStealTargets(playerIdx, cardIndex int) []int {
	args := m.Called(playerIdx, cardIndex)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockStealingBundlesGame) GetTableCards() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockStealingBundlesGame) GetPlayer(i int) *domain.StealingBundlesPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.StealingBundlesPlayer)
	}
	return nil
}

func (m *MockStealingBundlesGame) GetHint() *domain.StealingBundlesHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.StealingBundlesHint)
	}
	return nil
}

func (m *MockStealingBundlesGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

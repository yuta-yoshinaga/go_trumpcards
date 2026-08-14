//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRollingStoneGame ローリングストーンゲームモック
type MockRollingStoneGame struct {
	mock.Mock
}

func (m *MockRollingStoneGame) Reset()   { m.Called() }
func (m *MockRollingStoneGame) CpuPlay() { m.Called() }
func (m *MockRollingStoneGame) GiveUp()  { m.Called() }

func (m *MockRollingStoneGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockRollingStoneGame) PlayerPickUp() error { return m.Called().Error(0) }

func (m *MockRollingStoneGame) GetConfig() domain.RollingStoneConfig {
	return m.Called().Get(0).(domain.RollingStoneConfig)
}

func (m *MockRollingStoneGame) SetConfig(cfg domain.RollingStoneConfig) { m.Called(cfg) }

func (m *MockRollingStoneGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockRollingStoneGame) GetPhase() domain.RollingStonePhase {
	return m.Called().Get(0).(domain.RollingStonePhase)
}

func (m *MockRollingStoneGame) IsHumanTurn() bool             { return m.Called().Bool(0) }
func (m *MockRollingStoneGame) MustPickUp(playerIdx int) bool { return m.Called(playerIdx).Bool(0) }
func (m *MockRollingStoneGame) GetCurrentPlayerIdx() int      { return m.Called().Int(0) }
func (m *MockRollingStoneGame) GetLeadPlayerIdx() int         { return m.Called().Int(0) }
func (m *MockRollingStoneGame) GetTrickNumber() int           { return m.Called().Int(0) }
func (m *MockRollingStoneGame) GetLastPickupIdx() int         { return m.Called().Int(0) }
func (m *MockRollingStoneGame) GetFinishedCnt() int           { return m.Called().Int(0) }
func (m *MockRollingStoneGame) GetDiscarded() int             { return m.Called().Int(0) }
func (m *MockRollingStoneGame) GetDeckSize() int              { return m.Called().Int(0) }
func (m *MockRollingStoneGame) GetPlayerCnt() int             { return m.Called().Int(0) }
func (m *MockRollingStoneGame) GetWinnerIdx() int             { return m.Called().Int(0) }

func (m *MockRollingStoneGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockRollingStoneGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockRollingStoneGame) GetPlayer(i int) *domain.RollingStonePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.RollingStonePlayer)
	}
	return nil
}

func (m *MockRollingStoneGame) GetHint() *domain.RollingStoneHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.RollingStoneHint)
	}
	return nil
}

func (m *MockRollingStoneGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

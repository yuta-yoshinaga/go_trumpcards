//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBhabhiGame バービーゲームモック
type MockBhabhiGame struct {
	mock.Mock
}

func (m *MockBhabhiGame) Reset()   { m.Called() }
func (m *MockBhabhiGame) CpuPlay() { m.Called() }
func (m *MockBhabhiGame) GiveUp()  { m.Called() }

func (m *MockBhabhiGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockBhabhiGame) GetConfig() domain.BhabhiConfig {
	return m.Called().Get(0).(domain.BhabhiConfig)
}

func (m *MockBhabhiGame) SetConfig(cfg domain.BhabhiConfig) { m.Called(cfg) }

func (m *MockBhabhiGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockBhabhiGame) GetPhase() domain.BhabhiPhase {
	return m.Called().Get(0).(domain.BhabhiPhase)
}

func (m *MockBhabhiGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockBhabhiGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockBhabhiGame) GetLeadSuit() int         { return m.Called().Int(0) }
func (m *MockBhabhiGame) GetLastPickupIdx() int    { return m.Called().Int(0) }
func (m *MockBhabhiGame) GetLastPickupSize() int   { return m.Called().Int(0) }
func (m *MockBhabhiGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockBhabhiGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockBhabhiGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockBhabhiGame) GetAliveCount() int       { return m.Called().Int(0) }
func (m *MockBhabhiGame) GetBhabhiIdx() int        { return m.Called().Int(0) }
func (m *MockBhabhiGame) IsStalemate() bool        { return m.Called().Bool(0) }

func (m *MockBhabhiGame) GetPile() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockBhabhiGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockBhabhiGame) GetPlayer(i int) *domain.BhabhiPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.BhabhiPlayer)
	}
	return nil
}

func (m *MockBhabhiGame) GetHint() *domain.BhabhiHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.BhabhiHint)
	}
	return nil
}

func (m *MockBhabhiGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

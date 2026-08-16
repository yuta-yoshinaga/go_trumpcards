//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLingerLongerGame リンガーロンガーゲームモック
type MockLingerLongerGame struct {
	mock.Mock
}

func (m *MockLingerLongerGame) Reset()   { m.Called() }
func (m *MockLingerLongerGame) CpuPlay() { m.Called() }
func (m *MockLingerLongerGame) GiveUp()  { m.Called() }

func (m *MockLingerLongerGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockLingerLongerGame) GetConfig() domain.LingerLongerConfig {
	return m.Called().Get(0).(domain.LingerLongerConfig)
}

func (m *MockLingerLongerGame) SetConfig(cfg domain.LingerLongerConfig) { m.Called(cfg) }

func (m *MockLingerLongerGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockLingerLongerGame) GetPhase() domain.LingerLongerPhase {
	return m.Called().Get(0).(domain.LingerLongerPhase)
}

func (m *MockLingerLongerGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockLingerLongerGame) GetStockSize() int        { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetLastDrawIdx() int      { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetEliminatedCnt() int    { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetDiscarded() int        { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockLingerLongerGame) GetWinReason() string     { return m.Called().String(0) }

func (m *MockLingerLongerGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockLingerLongerGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockLingerLongerGame) GetPlayer(i int) *domain.LingerLongerPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.LingerLongerPlayer)
	}
	return nil
}

func (m *MockLingerLongerGame) GetHint() *domain.LingerLongerHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.LingerLongerHint)
	}
	return nil
}

func (m *MockLingerLongerGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

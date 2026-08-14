//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPigGame ピッグゲームモック
type MockPigGame struct {
	mock.Mock
}

func (m *MockPigGame) Reset()   { m.Called() }
func (m *MockPigGame) CpuPlay() { m.Called() }
func (m *MockPigGame) GiveUp()  { m.Called() }

func (m *MockPigGame) PlayerPass(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockPigGame) PlayerSignal() error            { return m.Called().Error(0) }
func (m *MockPigGame) NextRound() error               { return m.Called().Error(0) }

func (m *MockPigGame) GetConfig() domain.PigConfig {
	return m.Called().Get(0).(domain.PigConfig)
}

func (m *MockPigGame) SetConfig(cfg domain.PigConfig) { m.Called(cfg) }

func (m *MockPigGame) GetPhase() domain.PigPhase {
	return m.Called().Get(0).(domain.PigPhase)
}

func (m *MockPigGame) GetGameEndFlag() bool     { return m.Called().Bool(0) }
func (m *MockPigGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockPigGame) HasChosenPass(i int) bool { return m.Called(i).Bool(0) }
func (m *MockPigGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockPigGame) GetSignallerIdx() int     { return m.Called().Int(0) }
func (m *MockPigGame) GetNoticedCnt() int       { return m.Called().Int(0) }
func (m *MockPigGame) GetRoundLoserIdx() int    { return m.Called().Int(0) }
func (m *MockPigGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockPigGame) GetPassCount() int        { return m.Called().Int(0) }
func (m *MockPigGame) GetDeckSize() int         { return m.Called().Int(0) }
func (m *MockPigGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockPigGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockPigGame) GetValidPassIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockPigGame) GetPlayer(i int) *domain.PigPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.PigPlayer)
	}
	return nil
}

func (m *MockPigGame) GetHint() *domain.PigHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.PigHint)
	}
	return nil
}

func (m *MockPigGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMachiavelliGame マキャヴェッリゲームのモック
type MockMachiavelliGame struct {
	mock.Mock
}

func (m *MockMachiavelliGame) Reset()            { m.Called() }
func (m *MockMachiavelliGame) NextRound()        { m.Called() }
func (m *MockMachiavelliGame) PlayerDraw() error { return m.Called().Error(0) }
func (m *MockMachiavelliGame) PlayerPlay(refs [][]domain.MachiavelliCardRef, handIndices []int) error {
	return m.Called(refs, handIndices).Error(0)
}
func (m *MockMachiavelliGame) PlayerNewMeld(handIndices []int) error {
	return m.Called(handIndices).Error(0)
}
func (m *MockMachiavelliGame) PlayerLayoff(meldIdx, handIndex int) error {
	return m.Called(meldIdx, handIndex).Error(0)
}
func (m *MockMachiavelliGame) CpuPlay() { m.Called() }
func (m *MockMachiavelliGame) GetConfig() domain.MachiavelliConfig {
	return m.Called().Get(0).(domain.MachiavelliConfig)
}
func (m *MockMachiavelliGame) SetConfig(cfg domain.MachiavelliConfig) { m.Called(cfg) }
func (m *MockMachiavelliGame) GetGameEndFlag() bool                   { return m.Called().Bool(0) }
func (m *MockMachiavelliGame) GetPhase() domain.MachiavelliPhase {
	return m.Called().Get(0).(domain.MachiavelliPhase)
}
func (m *MockMachiavelliGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockMachiavelliGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockMachiavelliGame) GetTargetRounds() int     { return m.Called().Int(0) }
func (m *MockMachiavelliGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockMachiavelliGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockMachiavelliGame) GetTable() [][]*domain.Card {
	return m.Called().Get(0).([][]*domain.Card)
}
func (m *MockMachiavelliGame) GetDrawPileCount() int  { return m.Called().Int(0) }
func (m *MockMachiavelliGame) GetWinnerIdx() int      { return m.Called().Int(0) }
func (m *MockMachiavelliGame) GetRoundWinnerIdx() int { return m.Called().Int(0) }
func (m *MockMachiavelliGame) GetPlayerCnt() int      { return m.Called().Int(0) }
func (m *MockMachiavelliGame) GetPlayer(i int) *domain.MachiavelliPlayer {
	return m.Called(i).Get(0).(*domain.MachiavelliPlayer)
}
func (m *MockMachiavelliGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockMachiavelliGame) PlayerDeadwoodValue(i int) int { return m.Called(i).Int(0) }

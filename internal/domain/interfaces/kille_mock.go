//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKilleGame モック
type MockKilleGame struct {
	mock.Mock
}

func (m *MockKilleGame) Reset()                           { m.Called() }
func (m *MockKilleGame) NextRound() error                 { return m.Called().Error(0) }
func (m *MockKilleGame) Exchange(player int) error        { return m.Called(player).Error(0) }
func (m *MockKilleGame) Satisfied(player int) error       { return m.Called(player).Error(0) }
func (m *MockKilleGame) Reenter(seat int) error           { return m.Called(seat).Error(0) }
func (m *MockKilleGame) CpuPlay()                         { m.Called() }
func (m *MockKilleGame) KilleCpuReenterDecide(s int) bool { return m.Called(s).Bool(0) }
func (m *MockKilleGame) GetConfig() domain.KilleConfig {
	return m.Called().Get(0).(domain.KilleConfig)
}
func (m *MockKilleGame) SetConfig(cfg domain.KilleConfig) { m.Called(cfg) }
func (m *MockKilleGame) GetGameEndFlag() bool             { return m.Called().Bool(0) }
func (m *MockKilleGame) GetPhase() domain.KillePhase {
	return m.Called().Get(0).(domain.KillePhase)
}
func (m *MockKilleGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockKilleGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockKilleGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockKilleGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockKilleGame) GetStockCount() int       { return m.Called().Int(0) }
func (m *MockKilleGame) GetPot() int              { return m.Called().Int(0) }
func (m *MockKilleGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockKilleGame) GetPlayers() []*domain.KillePlayer {
	return m.Called().Get(0).([]*domain.KillePlayer)
}
func (m *MockKilleGame) GetPlayer(idx int) *domain.KillePlayer {
	return m.Called(idx).Get(0).(*domain.KillePlayer)
}
func (m *MockKilleGame) GetEvents() []*domain.KilleEvent {
	return m.Called().Get(0).([]*domain.KilleEvent)
}
func (m *MockKilleGame) GetLoserIdxs() []int           { return m.Called().Get(0).([]int) }
func (m *MockKilleGame) KilleStrength(seat int) int    { return m.Called(seat).Int(0) }
func (m *MockKilleGame) KilleReentryCost(seat int) int { return m.Called(seat).Int(0) }
func (m *MockKilleGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

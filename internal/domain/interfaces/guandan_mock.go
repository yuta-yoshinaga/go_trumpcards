//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGuandanGame モック
type MockGuandanGame struct {
	mock.Mock
}

func (m *MockGuandanGame) Reset()          { m.Called() }
func (m *MockGuandanGame) NextHand() error { return m.Called().Error(0) }
func (m *MockGuandanGame) PlayCards(player int, idxs []int) error {
	return m.Called(player, idxs).Error(0)
}
func (m *MockGuandanGame) Pass(player int) error { return m.Called(player).Error(0) }
func (m *MockGuandanGame) ReturnTribute(player, idx int) error {
	return m.Called(player, idx).Error(0)
}
func (m *MockGuandanGame) CpuPlay() { m.Called() }
func (m *MockGuandanGame) GetConfig() domain.GuandanConfig {
	return m.Called().Get(0).(domain.GuandanConfig)
}
func (m *MockGuandanGame) SetConfig(cfg domain.GuandanConfig) { m.Called(cfg) }
func (m *MockGuandanGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockGuandanGame) GetPhase() domain.GuandanPhase {
	return m.Called().Get(0).(domain.GuandanPhase)
}
func (m *MockGuandanGame) IsHumanTurn() bool         { return m.Called().Bool(0) }
func (m *MockGuandanGame) GetCurrentPlayerIdx() int  { return m.Called().Int(0) }
func (m *MockGuandanGame) GetLevel() int             { return m.Called().Int(0) }
func (m *MockGuandanGame) GetTeamLevel(team int) int { return m.Called(team).Int(0) }
func (m *MockGuandanGame) GetDeclarerTeam() int      { return m.Called().Int(0) }
func (m *MockGuandanGame) GetLastCombo() *domain.GuandanCombo {
	c, _ := m.Called().Get(0).(*domain.GuandanCombo)
	return c
}
func (m *MockGuandanGame) GetLastPlayerIdx() int { return m.Called().Int(0) }
func (m *MockGuandanGame) GetFinished() []int {
	return m.Called().Get(0).([]int)
}
func (m *MockGuandanGame) GetTributes() []*domain.GuandanTribute {
	return m.Called().Get(0).([]*domain.GuandanTribute)
}
func (m *MockGuandanGame) IsTributeCancelled() bool { return m.Called().Bool(0) }
func (m *MockGuandanGame) GetLastResult() *domain.GuandanHandResult {
	r, _ := m.Called().Get(0).(*domain.GuandanHandResult)
	return r
}
func (m *MockGuandanGame) GetHandNumber() int { return m.Called().Int(0) }
func (m *MockGuandanGame) GetWinnerTeam() int { return m.Called().Int(0) }
func (m *MockGuandanGame) GetPlayers() []*domain.GuandanPlayer {
	return m.Called().Get(0).([]*domain.GuandanPlayer)
}
func (m *MockGuandanGame) GetPlayer(idx int) *domain.GuandanPlayer {
	p, _ := m.Called(idx).Get(0).(*domain.GuandanPlayer)
	return p
}
func (m *MockGuandanGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

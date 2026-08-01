//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKarnoffelGame モック
type MockKarnoffelGame struct {
	mock.Mock
}

func (m *MockKarnoffelGame) Reset()          { m.Called() }
func (m *MockKarnoffelGame) NextHand() error { return m.Called().Error(0) }
func (m *MockKarnoffelGame) PlayCard(player, idx int) error {
	return m.Called(player, idx).Error(0)
}
func (m *MockKarnoffelGame) CpuPlay() { m.Called() }
func (m *MockKarnoffelGame) KarnoffelValidPlays(player int) []int {
	return m.Called(player).Get(0).([]int)
}
func (m *MockKarnoffelGame) KarnoffelTeamTricks(team int) int { return m.Called(team).Int(0) }
func (m *MockKarnoffelGame) GetConfig() domain.KarnoffelConfig {
	return m.Called().Get(0).(domain.KarnoffelConfig)
}
func (m *MockKarnoffelGame) SetConfig(cfg domain.KarnoffelConfig) { m.Called(cfg) }
func (m *MockKarnoffelGame) GetGameEndFlag() bool                 { return m.Called().Bool(0) }
func (m *MockKarnoffelGame) GetPhase() domain.KarnoffelPhase {
	return m.Called().Get(0).(domain.KarnoffelPhase)
}
func (m *MockKarnoffelGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockKarnoffelGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockKarnoffelGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockKarnoffelGame) GetChosenSuit() int       { return m.Called().Int(0) }
func (m *MockKarnoffelGame) GetUpCard(idx int) *domain.Card {
	c, _ := m.Called(idx).Get(0).(*domain.Card)
	return c
}
func (m *MockKarnoffelGame) GetTrick() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockKarnoffelGame) GetTrickLeaderIdx() int   { return m.Called().Int(0) }
func (m *MockKarnoffelGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockKarnoffelGame) GetTricksWon(idx int) int { return m.Called(idx).Int(0) }
func (m *MockKarnoffelGame) GetHandsWon(team int) int { return m.Called(team).Int(0) }
func (m *MockKarnoffelGame) GetLastResult() *domain.KarnoffelHandResult {
	r, _ := m.Called().Get(0).(*domain.KarnoffelHandResult)
	return r
}
func (m *MockKarnoffelGame) GetHandNumber() int { return m.Called().Int(0) }
func (m *MockKarnoffelGame) GetWinnerTeam() int { return m.Called().Int(0) }
func (m *MockKarnoffelGame) GetPlayers() []*domain.KarnoffelPlayer {
	return m.Called().Get(0).([]*domain.KarnoffelPlayer)
}
func (m *MockKarnoffelGame) GetPlayer(idx int) *domain.KarnoffelPlayer {
	p, _ := m.Called(idx).Get(0).(*domain.KarnoffelPlayer)
	return p
}
func (m *MockKarnoffelGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

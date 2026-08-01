//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLiteratureGame モック
type MockLiteratureGame struct {
	mock.Mock
}

func (m *MockLiteratureGame) Reset() { m.Called() }
func (m *MockLiteratureGame) Ask(from, to int, c *domain.Card) error {
	return m.Called(from, to, c).Error(0)
}
func (m *MockLiteratureGame) Claim(player, half int, holders []int) error {
	return m.Called(player, half, holders).Error(0)
}
func (m *MockLiteratureGame) CpuPlay() { m.Called() }
func (m *MockLiteratureGame) LiteratureCanAsk(from, to int, c *domain.Card) error {
	return m.Called(from, to, c).Error(0)
}
func (m *MockLiteratureGame) LiteratureHoldsHalfSuit(seat, half int) bool {
	return m.Called(seat, half).Bool(0)
}
func (m *MockLiteratureGame) LiteratureTeamHalfSuits(team int) int { return m.Called(team).Int(0) }
func (m *MockLiteratureGame) LiteratureCancelledCount() int        { return m.Called().Int(0) }
func (m *MockLiteratureGame) LiteratureOpenCount() int             { return m.Called().Int(0) }
func (m *MockLiteratureGame) GetConfig() domain.LiteratureConfig {
	return m.Called().Get(0).(domain.LiteratureConfig)
}
func (m *MockLiteratureGame) SetConfig(cfg domain.LiteratureConfig) { m.Called(cfg) }
func (m *MockLiteratureGame) GetGameEndFlag() bool                  { return m.Called().Bool(0) }
func (m *MockLiteratureGame) GetPhase() domain.LiteraturePhase {
	return m.Called().Get(0).(domain.LiteraturePhase)
}
func (m *MockLiteratureGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockLiteratureGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockLiteratureGame) GetHalfSuitState(half int) domain.LiteratureHalfSuitState {
	return m.Called(half).Get(0).(domain.LiteratureHalfSuitState)
}
func (m *MockLiteratureGame) GetAsks() []*domain.LiteratureAsk {
	return m.Called().Get(0).([]*domain.LiteratureAsk)
}
func (m *MockLiteratureGame) GetClaims() []*domain.LiteratureClaimResult {
	return m.Called().Get(0).([]*domain.LiteratureClaimResult)
}
func (m *MockLiteratureGame) GetLastAsk() *domain.LiteratureAsk {
	a, _ := m.Called().Get(0).(*domain.LiteratureAsk)
	return a
}
func (m *MockLiteratureGame) GetLastClaim() *domain.LiteratureClaimResult {
	c, _ := m.Called().Get(0).(*domain.LiteratureClaimResult)
	return c
}
func (m *MockLiteratureGame) GetWinnerTeam() int { return m.Called().Int(0) }
func (m *MockLiteratureGame) GetPlayers() []*domain.LiteraturePlayer {
	return m.Called().Get(0).([]*domain.LiteraturePlayer)
}
func (m *MockLiteratureGame) GetPlayer(idx int) *domain.LiteraturePlayer {
	p, _ := m.Called(idx).Get(0).(*domain.LiteraturePlayer)
	return p
}
func (m *MockLiteratureGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

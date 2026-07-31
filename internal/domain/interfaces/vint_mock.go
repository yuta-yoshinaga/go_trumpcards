//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockVintGame モック
type MockVintGame struct {
	mock.Mock
}

func (m *MockVintGame) Reset()          { m.Called() }
func (m *MockVintGame) NextHand() error { return m.Called().Error(0) }
func (m *MockVintGame) Bid(player, level, denom int) error {
	return m.Called(player, level, denom).Error(0)
}
func (m *MockVintGame) PassBid(player int) error       { return m.Called(player).Error(0) }
func (m *MockVintGame) PlayCard(player, idx int) error { return m.Called(player, idx).Error(0) }
func (m *MockVintGame) CpuPlay()                       { m.Called() }
func (m *MockVintGame) VintValidPlays(player int) []int {
	return m.Called(player).Get(0).([]int)
}
func (m *MockVintGame) VintTeamTricks(team int) int { return m.Called(team).Int(0) }
func (m *MockVintGame) GetConfig() domain.VintConfig {
	return m.Called().Get(0).(domain.VintConfig)
}
func (m *MockVintGame) SetConfig(cfg domain.VintConfig) { m.Called(cfg) }
func (m *MockVintGame) GetGameEndFlag() bool            { return m.Called().Bool(0) }
func (m *MockVintGame) GetPhase() domain.VintPhase {
	return m.Called().Get(0).(domain.VintPhase)
}
func (m *MockVintGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockVintGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockVintGame) GetBidPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockVintGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockVintGame) GetBids() []*domain.VintBid {
	return m.Called().Get(0).([]*domain.VintBid)
}
func (m *MockVintGame) GetHighBid() *domain.VintBid {
	b, _ := m.Called().Get(0).(*domain.VintBid)
	return b
}
func (m *MockVintGame) GetDeclarerIdx() int { return m.Called().Int(0) }
func (m *MockVintGame) GetTrumpSuit() int   { return m.Called().Int(0) }
func (m *MockVintGame) GetTrick() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockVintGame) GetTrickLeaderIdx() int   { return m.Called().Int(0) }
func (m *MockVintGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockVintGame) GetTricksWon(idx int) int { return m.Called(idx).Int(0) }
func (m *MockVintGame) GetBelow(team int) int    { return m.Called(team).Int(0) }
func (m *MockVintGame) GetAbove(team int) int    { return m.Called(team).Int(0) }
func (m *MockVintGame) GetGamesWon(team int) int { return m.Called(team).Int(0) }
func (m *MockVintGame) GetLastResult() *domain.VintHandResult {
	r, _ := m.Called().Get(0).(*domain.VintHandResult)
	return r
}
func (m *MockVintGame) GetHandNumber() int { return m.Called().Int(0) }
func (m *MockVintGame) GetWinnerTeam() int { return m.Called().Int(0) }
func (m *MockVintGame) GetPlayers() []*domain.VintPlayer {
	return m.Called().Get(0).([]*domain.VintPlayer)
}
func (m *MockVintGame) GetPlayer(idx int) *domain.VintPlayer {
	p, _ := m.Called(idx).Get(0).(*domain.VintPlayer)
	return p
}
func (m *MockVintGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

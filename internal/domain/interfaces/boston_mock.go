//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBostonGame モック
type MockBostonGame struct {
	mock.Mock
}

func (m *MockBostonGame) Reset()          { m.Called() }
func (m *MockBostonGame) NextHand() error { return m.Called().Error(0) }
func (m *MockBostonGame) Bid(player int, level domain.BostonBidLevel, suit int) error {
	return m.Called(player, level, suit).Error(0)
}
func (m *MockBostonGame) PassBid(player int) error { return m.Called(player).Error(0) }
func (m *MockBostonGame) CallPartner(player, partner int) error {
	return m.Called(player, partner).Error(0)
}
func (m *MockBostonGame) PlayCard(player, idx int) error { return m.Called(player, idx).Error(0) }
func (m *MockBostonGame) CpuPlay()                       { m.Called() }
func (m *MockBostonGame) BostonValidPlays(player int) []int {
	return m.Called(player).Get(0).([]int)
}
func (m *MockBostonGame) BostonIsDeclarerSide(seat int) bool { return m.Called(seat).Bool(0) }
func (m *MockBostonGame) BostonDeclarerTricks() int          { return m.Called().Int(0) }
func (m *MockBostonGame) GetConfig() domain.BostonConfig {
	return m.Called().Get(0).(domain.BostonConfig)
}
func (m *MockBostonGame) SetConfig(cfg domain.BostonConfig) { m.Called(cfg) }
func (m *MockBostonGame) GetGameEndFlag() bool              { return m.Called().Bool(0) }
func (m *MockBostonGame) GetPhase() domain.BostonPhase {
	return m.Called().Get(0).(domain.BostonPhase)
}
func (m *MockBostonGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockBostonGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockBostonGame) GetBidPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockBostonGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockBostonGame) GetBids() []*domain.BostonBidRecord {
	return m.Called().Get(0).([]*domain.BostonBidRecord)
}
func (m *MockBostonGame) GetHighBid() *domain.BostonBidRecord {
	b, _ := m.Called().Get(0).(*domain.BostonBidRecord)
	return b
}
func (m *MockBostonGame) GetDeclarerIdx() int { return m.Called().Int(0) }
func (m *MockBostonGame) GetPartnerIdx() int  { return m.Called().Int(0) }
func (m *MockBostonGame) GetTrumpSuit() int   { return m.Called().Int(0) }
func (m *MockBostonGame) IsExposed() bool     { return m.Called().Bool(0) }
func (m *MockBostonGame) GetTrick() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockBostonGame) GetTrickLeaderIdx() int   { return m.Called().Int(0) }
func (m *MockBostonGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockBostonGame) GetTricksWon(idx int) int { return m.Called(idx).Int(0) }
func (m *MockBostonGame) IsBidMade() bool          { return m.Called().Bool(0) }
func (m *MockBostonGame) GetChips(idx int) int     { return m.Called(idx).Int(0) }
func (m *MockBostonGame) GetHandNumber() int       { return m.Called().Int(0) }
func (m *MockBostonGame) GetTargetHands() int      { return m.Called().Int(0) }
func (m *MockBostonGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockBostonGame) GetPlayers() []*domain.BostonPlayer {
	return m.Called().Get(0).([]*domain.BostonPlayer)
}
func (m *MockBostonGame) GetPlayer(idx int) *domain.BostonPlayer {
	p, _ := m.Called(idx).Get(0).(*domain.BostonPlayer)
	return p
}
func (m *MockBostonGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

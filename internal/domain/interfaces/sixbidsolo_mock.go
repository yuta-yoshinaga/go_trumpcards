//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSixBidSoloGame モック
type MockSixBidSoloGame struct {
	mock.Mock
}

func (m *MockSixBidSoloGame) Reset()          { m.Called() }
func (m *MockSixBidSoloGame) NextHand() error { return m.Called().Error(0) }
func (m *MockSixBidSoloGame) Bid(player int, kind domain.SixBidSoloBidKind) error {
	return m.Called(player, kind).Error(0)
}
func (m *MockSixBidSoloGame) PassBid(player int) error { return m.Called(player).Error(0) }
func (m *MockSixBidSoloGame) Declare(player, suit int, called *domain.Card) error {
	return m.Called(player, suit, called).Error(0)
}
func (m *MockSixBidSoloGame) PlayCard(player, idx int) error {
	return m.Called(player, idx).Error(0)
}
func (m *MockSixBidSoloGame) CpuPlay() { m.Called() }
func (m *MockSixBidSoloGame) SixBidSoloValidPlays(player int) []int {
	return m.Called(player).Get(0).([]int)
}
func (m *MockSixBidSoloGame) SixBidSoloWidowPoints() int { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetConfig() domain.SixBidSoloConfig {
	return m.Called().Get(0).(domain.SixBidSoloConfig)
}
func (m *MockSixBidSoloGame) SetConfig(cfg domain.SixBidSoloConfig) { m.Called(cfg) }
func (m *MockSixBidSoloGame) GetGameEndFlag() bool                  { return m.Called().Bool(0) }
func (m *MockSixBidSoloGame) GetPhase() domain.SixBidSoloPhase {
	return m.Called().Get(0).(domain.SixBidSoloPhase)
}
func (m *MockSixBidSoloGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockSixBidSoloGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetBidPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetBids() []*domain.SixBidSoloBid {
	return m.Called().Get(0).([]*domain.SixBidSoloBid)
}
func (m *MockSixBidSoloGame) GetHighBid() *domain.SixBidSoloBid {
	b, _ := m.Called().Get(0).(*domain.SixBidSoloBid)
	return b
}
func (m *MockSixBidSoloGame) GetDeclarerIdx() int { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetTrumpSuit() int   { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) IsDeclared() bool    { return m.Called().Bool(0) }
func (m *MockSixBidSoloGame) GetCalledCard() *domain.Card {
	c, _ := m.Called().Get(0).(*domain.Card)
	return c
}
func (m *MockSixBidSoloGame) IsSpreadOpen() bool { return m.Called().Bool(0) }
func (m *MockSixBidSoloGame) GetWidow() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockSixBidSoloGame) GetTrick() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockSixBidSoloGame) GetTrickLeaderIdx() int   { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetPoints(idx int) int    { return m.Called(idx).Int(0) }
func (m *MockSixBidSoloGame) GetTricksWon(idx int) int { return m.Called(idx).Int(0) }
func (m *MockSixBidSoloGame) GetScore(idx int) int     { return m.Called(idx).Int(0) }
func (m *MockSixBidSoloGame) GetLastResult() *domain.SixBidSoloHandResult {
	r, _ := m.Called().Get(0).(*domain.SixBidSoloHandResult)
	return r
}
func (m *MockSixBidSoloGame) GetHandNumber() int { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetWinnerIdx() int  { return m.Called().Int(0) }
func (m *MockSixBidSoloGame) GetPlayers() []*domain.SixBidSoloPlayer {
	return m.Called().Get(0).([]*domain.SixBidSoloPlayer)
}
func (m *MockSixBidSoloGame) GetPlayer(idx int) *domain.SixBidSoloPlayer {
	p, _ := m.Called(idx).Get(0).(*domain.SixBidSoloPlayer)
	return p
}
func (m *MockSixBidSoloGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

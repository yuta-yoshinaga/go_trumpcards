//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBidEuchreGame モック
type MockBidEuchreGame struct {
	mock.Mock
}

func (m *MockBidEuchreGame) Reset()                      { m.Called() }
func (m *MockBidEuchreGame) NextHand() error             { return m.Called().Error(0) }
func (m *MockBidEuchreGame) Bid(player, value int) error { return m.Called(player, value).Error(0) }
func (m *MockBidEuchreGame) PassBid(player int) error    { return m.Called(player).Error(0) }
func (m *MockBidEuchreGame) PlayCard(player, idx int) error {
	return m.Called(player, idx).Error(0)
}
func (m *MockBidEuchreGame) ChooseTrump(player int, t domain.BidEuchreTrump) error {
	return m.Called(player, t).Error(0)
}
func (m *MockBidEuchreGame) CpuPlay() { m.Called() }
func (m *MockBidEuchreGame) BidEuchreValidPlays(player int) []int {
	return m.Called(player).Get(0).([]int)
}
func (m *MockBidEuchreGame) BidEuchreTeamTricks(team int) int { return m.Called(team).Int(0) }
func (m *MockBidEuchreGame) GetConfig() domain.BidEuchreConfig {
	return m.Called().Get(0).(domain.BidEuchreConfig)
}
func (m *MockBidEuchreGame) SetConfig(cfg domain.BidEuchreConfig) { m.Called(cfg) }
func (m *MockBidEuchreGame) GetGameEndFlag() bool                 { return m.Called().Bool(0) }
func (m *MockBidEuchreGame) GetPhase() domain.BidEuchrePhase {
	return m.Called().Get(0).(domain.BidEuchrePhase)
}
func (m *MockBidEuchreGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockBidEuchreGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockBidEuchreGame) GetBidPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockBidEuchreGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockBidEuchreGame) GetBids() []*domain.BidEuchreBid {
	return m.Called().Get(0).([]*domain.BidEuchreBid)
}
func (m *MockBidEuchreGame) GetHighBid() *domain.BidEuchreBid {
	b, _ := m.Called().Get(0).(*domain.BidEuchreBid)
	return b
}
func (m *MockBidEuchreGame) GetDeclarerIdx() int { return m.Called().Int(0) }
func (m *MockBidEuchreGame) GetTrump() domain.BidEuchreTrump {
	return m.Called().Get(0).(domain.BidEuchreTrump)
}
func (m *MockBidEuchreGame) GetTrumpSuit() int   { return m.Called().Int(0) }
func (m *MockBidEuchreGame) IsTrumpChosen() bool { return m.Called().Bool(0) }
func (m *MockBidEuchreGame) GetTrick() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockBidEuchreGame) GetTrickLeaderIdx() int   { return m.Called().Int(0) }
func (m *MockBidEuchreGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockBidEuchreGame) GetTricksWon(idx int) int { return m.Called(idx).Int(0) }
func (m *MockBidEuchreGame) GetScore(team int) int    { return m.Called(team).Int(0) }
func (m *MockBidEuchreGame) GetLastResult() *domain.BidEuchreHandResult {
	r, _ := m.Called().Get(0).(*domain.BidEuchreHandResult)
	return r
}
func (m *MockBidEuchreGame) GetHandNumber() int { return m.Called().Int(0) }
func (m *MockBidEuchreGame) GetWinnerTeam() int { return m.Called().Int(0) }
func (m *MockBidEuchreGame) GetPlayers() []*domain.BidEuchrePlayer {
	return m.Called().Get(0).([]*domain.BidEuchrePlayer)
}
func (m *MockBidEuchreGame) GetPlayer(idx int) *domain.BidEuchrePlayer {
	p, _ := m.Called(idx).Get(0).(*domain.BidEuchrePlayer)
	return p
}
func (m *MockBidEuchreGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

// BidEuchreMinLegalBid モック
func (m *MockBidEuchreGame) BidEuchreMinLegalBid(player int) (int, bool) {
	ret := m.Called(player)
	return ret.Int(0), ret.Bool(1)
}

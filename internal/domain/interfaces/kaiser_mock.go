//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKaiserGame モック
type MockKaiserGame struct {
	mock.Mock
}

func (m *MockKaiserGame) Reset()          { m.Called() }
func (m *MockKaiserGame) NextHand() error { return m.Called().Error(0) }
func (m *MockKaiserGame) Bid(player, value int, contract domain.KaiserContract) error {
	return m.Called(player, value, contract).Error(0)
}
func (m *MockKaiserGame) PassBid(player int) error        { return m.Called(player).Error(0) }
func (m *MockKaiserGame) SetTrump(player, suit int) error { return m.Called(player, suit).Error(0) }
func (m *MockKaiserGame) Discard(player int, idxs []int) error {
	return m.Called(player, idxs).Error(0)
}
func (m *MockKaiserGame) PlayCard(player, idx int) error { return m.Called(player, idx).Error(0) }
func (m *MockKaiserGame) CpuPlay()                       { m.Called() }
func (m *MockKaiserGame) KaiserValidPlays(player int) []int {
	return m.Called(player).Get(0).([]int)
}
func (m *MockKaiserGame) GetConfig() domain.KaiserConfig {
	return m.Called().Get(0).(domain.KaiserConfig)
}
func (m *MockKaiserGame) SetConfig(cfg domain.KaiserConfig) { m.Called(cfg) }
func (m *MockKaiserGame) GetGameEndFlag() bool              { return m.Called().Bool(0) }
func (m *MockKaiserGame) GetPhase() domain.KaiserPhase {
	return m.Called().Get(0).(domain.KaiserPhase)
}
func (m *MockKaiserGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockKaiserGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockKaiserGame) GetBidPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockKaiserGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockKaiserGame) GetBids() []*domain.KaiserBid {
	return m.Called().Get(0).([]*domain.KaiserBid)
}
func (m *MockKaiserGame) GetHighBid() *domain.KaiserBid {
	b, _ := m.Called().Get(0).(*domain.KaiserBid)
	return b
}
func (m *MockKaiserGame) GetDeclarerIdx() int { return m.Called().Int(0) }
func (m *MockKaiserGame) GetTrumpSuit() int   { return m.Called().Int(0) }
func (m *MockKaiserGame) GetContract() domain.KaiserContract {
	return m.Called().Get(0).(domain.KaiserContract)
}
func (m *MockKaiserGame) GetKittySize() int { return m.Called().Int(0) }
func (m *MockKaiserGame) GetTrick() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockKaiserGame) GetTrickLeaderIdx() int     { return m.Called().Int(0) }
func (m *MockKaiserGame) GetTrickNumber() int        { return m.Called().Int(0) }
func (m *MockKaiserGame) GetHandPoints(team int) int { return m.Called(team).Int(0) }
func (m *MockKaiserGame) GetHeartFiveBy() int        { return m.Called().Int(0) }
func (m *MockKaiserGame) GetSpadeThreeBy() int       { return m.Called().Int(0) }
func (m *MockKaiserGame) IsBidMade() bool            { return m.Called().Bool(0) }
func (m *MockKaiserGame) GetScore(team int) int      { return m.Called(team).Int(0) }
func (m *MockKaiserGame) GetTargetScore() int        { return m.Called().Int(0) }
func (m *MockKaiserGame) GetHandNumber() int         { return m.Called().Int(0) }
func (m *MockKaiserGame) GetWinnerTeam() int         { return m.Called().Int(0) }
func (m *MockKaiserGame) GetPlayers() []*domain.KaiserPlayer {
	return m.Called().Get(0).([]*domain.KaiserPlayer)
}
func (m *MockKaiserGame) GetPlayer(idx int) *domain.KaiserPlayer {
	p, _ := m.Called(idx).Get(0).(*domain.KaiserPlayer)
	return p
}
func (m *MockKaiserGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}

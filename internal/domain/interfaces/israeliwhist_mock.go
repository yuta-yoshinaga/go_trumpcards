//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockIsraeliWhistGame イスラエリホイストゲームモック
type MockIsraeliWhistGame struct {
	mock.Mock
}

func (m *MockIsraeliWhistGame) Reset()      { m.Called() }
func (m *MockIsraeliWhistGame) CpuAuction() { m.Called() }
func (m *MockIsraeliWhistGame) CpuBid()     { m.Called() }
func (m *MockIsraeliWhistGame) CpuPlay()    { m.Called() }
func (m *MockIsraeliWhistGame) NextRound()  { m.Called() }
func (m *MockIsraeliWhistGame) GiveUp()     { m.Called() }

func (m *MockIsraeliWhistGame) PlayerAuctionBid(bid, suit int) error {
	return m.Called(bid, suit).Error(0)
}
func (m *MockIsraeliWhistGame) PlayerAuctionPass() error { return m.Called().Error(0) }
func (m *MockIsraeliWhistGame) PlayerBid(bid int) error  { return m.Called(bid).Error(0) }

func (m *MockIsraeliWhistGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockIsraeliWhistGame) GetConfig() domain.IsraeliWhistConfig {
	return m.Called().Get(0).(domain.IsraeliWhistConfig)
}

func (m *MockIsraeliWhistGame) SetConfig(cfg domain.IsraeliWhistConfig) { m.Called(cfg) }

func (m *MockIsraeliWhistGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockIsraeliWhistGame) GetPhase() domain.IsraeliWhistPhase {
	return m.Called().Get(0).(domain.IsraeliWhistPhase)
}

func (m *MockIsraeliWhistGame) IsHumanTurn() bool         { return m.Called().Bool(0) }
func (m *MockIsraeliWhistGame) IsHumanAuctionTurn() bool  { return m.Called().Bool(0) }
func (m *MockIsraeliWhistGame) IsHumanBidTurn() bool      { return m.Called().Bool(0) }
func (m *MockIsraeliWhistGame) MinimumBidFor(idx int) int { return m.Called(idx).Int(0) }
func (m *MockIsraeliWhistGame) GetRestrictedBid() int     { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetRoundNumber() int       { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetTrickNumber() int       { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetTrumpSuit() int         { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetDeclarerIdx() int       { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetHighBid() int           { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetHighSuit() int          { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetCurrentPlayerIdx() int  { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetAuctionPlayerIdx() int  { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetBidPlayerIdx() int      { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetLeadPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetDealerIdx() int         { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetPlayerCnt() int         { return m.Called().Int(0) }
func (m *MockIsraeliWhistGame) GetWinnerIdx() int         { return m.Called().Int(0) }

func (m *MockIsraeliWhistGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockIsraeliWhistGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockIsraeliWhistGame) GetPlayer(i int) *domain.IsraeliWhistPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.IsraeliWhistPlayer)
}

func (m *MockIsraeliWhistGame) GetHint() *domain.IsraeliWhistHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.IsraeliWhistHint)
}

func (m *MockIsraeliWhistGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

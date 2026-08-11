//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHasenpfefferGame ハーゼンプフェファーゲームモック
type MockHasenpfefferGame struct {
	mock.Mock
}

func (m *MockHasenpfefferGame) Reset()      { m.Called() }
func (m *MockHasenpfefferGame) CpuBid()     { m.Called() }
func (m *MockHasenpfefferGame) CpuDiscard() { m.Called() }
func (m *MockHasenpfefferGame) CpuPlay()    { m.Called() }
func (m *MockHasenpfefferGame) NextHand()   { m.Called() }
func (m *MockHasenpfefferGame) GiveUp()     { m.Called() }

func (m *MockHasenpfefferGame) PlayerBid(bid int) error { return m.Called(bid).Error(0) }

func (m *MockHasenpfefferGame) PlayerDiscard(cardIndex, suit int) error {
	return m.Called(cardIndex, suit).Error(0)
}

func (m *MockHasenpfefferGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockHasenpfefferGame) GetConfig() domain.HasenpfefferConfig {
	return m.Called().Get(0).(domain.HasenpfefferConfig)
}

func (m *MockHasenpfefferGame) SetConfig(cfg domain.HasenpfefferConfig) { m.Called(cfg) }

func (m *MockHasenpfefferGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockHasenpfefferGame) GetPhase() domain.HasenpfefferPhase {
	return m.Called().Get(0).(domain.HasenpfefferPhase)
}

func (m *MockHasenpfefferGame) IsHumanTurn() bool          { return m.Called().Bool(0) }
func (m *MockHasenpfefferGame) IsHumanBidTurn() bool       { return m.Called().Bool(0) }
func (m *MockHasenpfefferGame) IsHumanDiscardTurn() bool   { return m.Called().Bool(0) }
func (m *MockHasenpfefferGame) MustBid(playerIdx int) bool { return m.Called(playerIdx).Bool(0) }
func (m *MockHasenpfefferGame) NextBid() int               { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetHandNumber() int         { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetTrickNumber() int        { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetTrumpSuit() int          { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetDeclarerIdx() int        { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetContract() int           { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetBlindSize() int          { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetCurrentPlayerIdx() int   { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetLeadPlayerIdx() int      { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetDealerIdx() int          { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetScore(team int) int      { return m.Called(team).Int(0) }
func (m *MockHasenpfefferGame) TeamTricks(team int) int    { return m.Called(team).Int(0) }
func (m *MockHasenpfefferGame) GetLastHandEuchred() bool   { return m.Called().Bool(0) }
func (m *MockHasenpfefferGame) GetLastHandTricks() int     { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetPlayerCnt() int          { return m.Called().Int(0) }
func (m *MockHasenpfefferGame) GetWinnerTeam() int         { return m.Called().Int(0) }

func (m *MockHasenpfefferGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockHasenpfefferGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockHasenpfefferGame) GetPlayer(i int) *domain.HasenpfefferPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.HasenpfefferPlayer)
	}
	return nil
}

func (m *MockHasenpfefferGame) GetHint() *domain.HasenpfefferHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.HasenpfefferHint)
	}
	return nil
}

func (m *MockHasenpfefferGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

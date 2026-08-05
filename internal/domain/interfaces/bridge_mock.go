//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBridgeGame ブリッジゲームモック
type MockBridgeGame struct {
	mock.Mock
}

func (m *MockBridgeGame) Reset()     { m.Called() }
func (m *MockBridgeGame) NextRound() { m.Called() }

func (m *MockBridgeGame) PlayerBid(bidType int, level int, suit int) error {
	args := m.Called(bidType, level, suit)
	return args.Error(0)
}

func (m *MockBridgeGame) CpuBid() { m.Called() }

func (m *MockBridgeGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBridgeGame) CpuPlay()      { m.Called() }
func (m *MockBridgeGame) ResolveTrick() { m.Called() }
func (m *MockBridgeGame) NextTrick()    { m.Called() }
func (m *MockBridgeGame) ScoreRound()   { m.Called() }

func (m *MockBridgeGame) GetConfig() domain.BridgeConfig {
	args := m.Called()
	return args.Get(0).(domain.BridgeConfig)
}

func (m *MockBridgeGame) SetConfig(cfg domain.BridgeConfig) { m.Called(cfg) }

func (m *MockBridgeGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockBridgeGame) GetPhase() domain.BridgePhase {
	args := m.Called()
	return args.Get(0).(domain.BridgePhase)
}

func (m *MockBridgeGame) IsHumanTurn() bool    { return m.Called().Bool(0) }
func (m *MockBridgeGame) IsHumanBidTurn() bool { return m.Called().Bool(0) }
func (m *MockBridgeGame) GetRoundNumber() int  { return m.Called().Int(0) }
func (m *MockBridgeGame) GetTrickNumber() int  { return m.Called().Int(0) }
func (m *MockBridgeGame) GetCurrentPlayerIdx() int {
	return m.Called().Int(0)
}

func (m *MockBridgeGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockBridgeGame) GetLeadPlayerIdx() int { return m.Called().Int(0) }
func (m *MockBridgeGame) GetBidPlayerIdx() int  { return m.Called().Int(0) }
func (m *MockBridgeGame) GetDealerIdx() int     { return m.Called().Int(0) }
func (m *MockBridgeGame) GetTrumpSuit() int     { return m.Called().Int(0) }
func (m *MockBridgeGame) GetContractLevel() int { return m.Called().Int(0) }
func (m *MockBridgeGame) GetContractSuit() int  { return m.Called().Int(0) }
func (m *MockBridgeGame) GetDoubled() int       { return m.Called().Int(0) }
func (m *MockBridgeGame) GetDeclarerIdx() int   { return m.Called().Int(0) }
func (m *MockBridgeGame) GetDummyIdx() int      { return m.Called().Int(0) }

func (m *MockBridgeGame) GetBidHistory() []*domain.BridgeBidEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.BridgeBidEntry)
	}
	return nil
}

func (m *MockBridgeGame) GetVulnerability(team int) bool {
	return m.Called(team).Bool(0)
}

func (m *MockBridgeGame) GetTeamScore(team int) int {
	return m.Called(team).Int(0)
}

func (m *MockBridgeGame) GetGamesWon(team int) int {
	return m.Called(team).Int(0)
}

func (m *MockBridgeGame) GetBelowLine(team int) int {
	return m.Called(team).Int(0)
}

func (m *MockBridgeGame) GetWinnerTeam() int { return m.Called().Int(0) }
func (m *MockBridgeGame) GetPlayerCnt() int  { return m.Called().Int(0) }

func (m *MockBridgeGame) GetPlayer(i int) *domain.BridgePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.BridgePlayer)
	}
	return nil
}

func (m *MockBridgeGame) IsOpeningLeadDone() bool { return m.Called().Bool(0) }

func (m *MockBridgeGame) GetDummyHand() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockBridgeGame) GetHint() *domain.BridgeHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.BridgeHint)
	}
	return nil
}

func (m *MockBridgeGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockBridgeGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

// BridgeMinLegalBid モック
func (m *MockBridgeGame) BridgeMinLegalBid() (int, int, bool) {
	ret := m.Called()
	return ret.Int(0), ret.Int(1), ret.Bool(2)
}

// BridgeCanDouble モック
func (m *MockBridgeGame) BridgeCanDouble(playerIdx int) bool { return m.Called(playerIdx).Bool(0) }

// BridgeCanRedouble モック
func (m *MockBridgeGame) BridgeCanRedouble(playerIdx int) bool { return m.Called(playerIdx).Bool(0) }

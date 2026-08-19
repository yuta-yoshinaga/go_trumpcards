//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHoneymoonBridgeGame ハネムーンブリッジゲームモック
type MockHoneymoonBridgeGame struct {
	mock.Mock
}

func (m *MockHoneymoonBridgeGame) Reset()     { m.Called() }
func (m *MockHoneymoonBridgeGame) CpuPlay()   { m.Called() }
func (m *MockHoneymoonBridgeGame) CpuBid()    { m.Called() }
func (m *MockHoneymoonBridgeGame) NextRound() { m.Called() }
func (m *MockHoneymoonBridgeGame) GiveUp()    { m.Called() }

func (m *MockHoneymoonBridgeGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockHoneymoonBridgeGame) PlayerBid(level, suit int) error {
	return m.Called(level, suit).Error(0)
}

func (m *MockHoneymoonBridgeGame) PlayerPass() error { return m.Called().Error(0) }

func (m *MockHoneymoonBridgeGame) GetConfig() domain.HoneymoonBridgeConfig {
	return m.Called().Get(0).(domain.HoneymoonBridgeConfig)
}

func (m *MockHoneymoonBridgeGame) SetConfig(cfg domain.HoneymoonBridgeConfig) { m.Called(cfg) }

func (m *MockHoneymoonBridgeGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockHoneymoonBridgeGame) GetPhase() domain.HoneymoonBridgePhase {
	return m.Called().Get(0).(domain.HoneymoonBridgePhase)
}

func (m *MockHoneymoonBridgeGame) NextBid() (int, int) {
	args := m.Called()
	return args.Int(0), args.Int(1)
}

func (m *MockHoneymoonBridgeGame) IsHumanTurn() bool     { return m.Called().Bool(0) }
func (m *MockHoneymoonBridgeGame) IsHumanBidTurn() bool  { return m.Called().Bool(0) }
func (m *MockHoneymoonBridgeGame) GetRoundNumber() int   { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetTrickNumber() int   { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetStockSize() int     { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetTrumpSuit() int     { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetDeclarerIdx() int   { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetContractLevel() int { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) RequiredTricks() int   { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetLastMade() bool     { return m.Called().Bool(0) }
func (m *MockHoneymoonBridgeGame) GetLastTricks() int    { return m.Called().Int(0) }

// GetLastPoints モック
func (m *MockHoneymoonBridgeGame) GetLastPoints() int       { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockHoneymoonBridgeGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockHoneymoonBridgeGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockHoneymoonBridgeGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockHoneymoonBridgeGame) GetPlayer(i int) *domain.HoneymoonBridgePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.HoneymoonBridgePlayer)
	}
	return nil
}

func (m *MockHoneymoonBridgeGame) GetHint() *domain.HoneymoonBridgeHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.HoneymoonBridgeHint)
	}
	return nil
}

func (m *MockHoneymoonBridgeGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

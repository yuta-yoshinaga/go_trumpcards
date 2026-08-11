//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEstimationGame エスティメーションゲームモック
type MockEstimationGame struct {
	mock.Mock
}

func (m *MockEstimationGame) Reset()          { m.Called() }
func (m *MockEstimationGame) CpuSelectTrump() { m.Called() }
func (m *MockEstimationGame) CpuBid()         { m.Called() }
func (m *MockEstimationGame) CpuPlay()        { m.Called() }
func (m *MockEstimationGame) NextRound()      { m.Called() }
func (m *MockEstimationGame) GiveUp()         { m.Called() }

func (m *MockEstimationGame) SelectTrump(suit int) error { return m.Called(suit).Error(0) }
func (m *MockEstimationGame) PlayerBid(bid int) error    { return m.Called(bid).Error(0) }

func (m *MockEstimationGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockEstimationGame) GetConfig() domain.EstimationConfig {
	return m.Called().Get(0).(domain.EstimationConfig)
}

func (m *MockEstimationGame) SetConfig(cfg domain.EstimationConfig) { m.Called(cfg) }

func (m *MockEstimationGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockEstimationGame) GetPhase() domain.EstimationPhase {
	return m.Called().Get(0).(domain.EstimationPhase)
}

func (m *MockEstimationGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockEstimationGame) IsHumanBidTurn() bool     { return m.Called().Bool(0) }
func (m *MockEstimationGame) IsHumanTrumpTurn() bool   { return m.Called().Bool(0) }
func (m *MockEstimationGame) GetRestrictedBid() int    { return m.Called().Int(0) }
func (m *MockEstimationGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockEstimationGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockEstimationGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockEstimationGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockEstimationGame) GetBidPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockEstimationGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockEstimationGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockEstimationGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockEstimationGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockEstimationGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockEstimationGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockEstimationGame) GetPlayer(i int) *domain.EstimationPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.EstimationPlayer)
}

func (m *MockEstimationGame) GetHint() *domain.EstimationHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.EstimationHint)
}

func (m *MockEstimationGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

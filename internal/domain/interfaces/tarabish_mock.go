//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTarabishGame タラビッシュゲームモック
type MockTarabishGame struct {
	mock.Mock
}

func (m *MockTarabishGame) Reset()     { m.Called() }
func (m *MockTarabishGame) CpuBid()    { m.Called() }
func (m *MockTarabishGame) CpuPlay()   { m.Called() }
func (m *MockTarabishGame) NextRound() { m.Called() }
func (m *MockTarabishGame) GiveUp()    { m.Called() }

func (m *MockTarabishGame) TakeTrump() error { return m.Called().Error(0) }
func (m *MockTarabishGame) PassTrump() error { return m.Called().Error(0) }

func (m *MockTarabishGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockTarabishGame) GetConfig() domain.TarabishConfig {
	return m.Called().Get(0).(domain.TarabishConfig)
}

func (m *MockTarabishGame) SetConfig(cfg domain.TarabishConfig) { m.Called(cfg) }

func (m *MockTarabishGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockTarabishGame) GetPhase() domain.TarabishPhase {
	return m.Called().Get(0).(domain.TarabishPhase)
}

func (m *MockTarabishGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockTarabishGame) IsHumanBidTurn() bool     { return m.Called().Bool(0) }
func (m *MockTarabishGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockTarabishGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockTarabishGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockTarabishGame) GetTrumpTakerIdx() int    { return m.Called().Int(0) }
func (m *MockTarabishGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockTarabishGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockTarabishGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockTarabishGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockTarabishGame) GetWinnerTeam() int       { return m.Called().Int(0) }

func (m *MockTarabishGame) GetScore(team int) int       { return m.Called(team).Int(0) }
func (m *MockTarabishGame) GetRoundPoints(team int) int { return m.Called(team).Int(0) }

func (m *MockTarabishGame) GetUpCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockTarabishGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockTarabishGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockTarabishGame) GetPlayer(i int) *domain.TarabishPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.TarabishPlayer)
}

func (m *MockTarabishGame) GetHint() *domain.TarabishHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.TarabishHint)
}

func (m *MockTarabishGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

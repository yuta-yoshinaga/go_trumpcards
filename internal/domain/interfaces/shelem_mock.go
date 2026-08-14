//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShelemGame シェレムゲームモック
type MockShelemGame struct {
	mock.Mock
}

func (m *MockShelemGame) Reset()     { m.Called() }
func (m *MockShelemGame) CpuBid()    { m.Called() }
func (m *MockShelemGame) CpuPlay()   { m.Called() }
func (m *MockShelemGame) NextRound() { m.Called() }
func (m *MockShelemGame) GiveUp()    { m.Called() }

func (m *MockShelemGame) PlayerBid(bid int) error { return m.Called(bid).Error(0) }
func (m *MockShelemGame) PlayerBidShelem() error  { return m.Called().Error(0) }
func (m *MockShelemGame) PlayerPass() error       { return m.Called().Error(0) }

func (m *MockShelemGame) PlayerDiscard(indices []int, suit int) error {
	return m.Called(indices, suit).Error(0)
}

func (m *MockShelemGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockShelemGame) GetConfig() domain.ShelemConfig {
	return m.Called().Get(0).(domain.ShelemConfig)
}

func (m *MockShelemGame) SetConfig(cfg domain.ShelemConfig) { m.Called(cfg) }

func (m *MockShelemGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockShelemGame) GetPhase() domain.ShelemPhase {
	return m.Called().Get(0).(domain.ShelemPhase)
}

func (m *MockShelemGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockShelemGame) IsHumanBidTurn() bool     { return m.Called().Bool(0) }
func (m *MockShelemGame) IsHumanDiscardTurn() bool { return m.Called().Bool(0) }
func (m *MockShelemGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockShelemGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockShelemGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockShelemGame) GetDeclarerIdx() int      { return m.Called().Int(0) }
func (m *MockShelemGame) GetContract() int         { return m.Called().Int(0) }
func (m *MockShelemGame) GetShelemBid() bool       { return m.Called().Bool(0) }
func (m *MockShelemGame) GetWidowSize() int        { return m.Called().Int(0) }
func (m *MockShelemGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockShelemGame) GetBidPlayerIdx() int     { return m.Called().Int(0) }
func (m *MockShelemGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockShelemGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockShelemGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockShelemGame) GetWinnerTeam() int       { return m.Called().Int(0) }

func (m *MockShelemGame) GetScore(team int) int       { return m.Called(team).Int(0) }
func (m *MockShelemGame) GetRoundPoints(team int) int { return m.Called(team).Int(0) }
func (m *MockShelemGame) TeamTricks(team int) int     { return m.Called(team).Int(0) }

func (m *MockShelemGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockShelemGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockShelemGame) GetPlayer(i int) *domain.ShelemPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ShelemPlayer)
}

func (m *MockShelemGame) GetHint() *domain.ShelemHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ShelemHint)
}

func (m *MockShelemGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

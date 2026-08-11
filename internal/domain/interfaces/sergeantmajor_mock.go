//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSergeantMajorGame サージェントメジャーゲームモック
type MockSergeantMajorGame struct {
	mock.Mock
}

func (m *MockSergeantMajorGame) Reset()           { m.Called() }
func (m *MockSergeantMajorGame) CpuDeclareTrump() { m.Called() }
func (m *MockSergeantMajorGame) CpuDiscard()      { m.Called() }
func (m *MockSergeantMajorGame) CpuPlay()         { m.Called() }
func (m *MockSergeantMajorGame) NextRound()       { m.Called() }
func (m *MockSergeantMajorGame) GiveUp()          { m.Called() }

func (m *MockSergeantMajorGame) PlayerDeclareTrump(suit int) error {
	return m.Called(suit).Error(0)
}

func (m *MockSergeantMajorGame) PlayerDiscard(indices []int) error {
	return m.Called(indices).Error(0)
}

func (m *MockSergeantMajorGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockSergeantMajorGame) GetConfig() domain.SergeantMajorConfig {
	return m.Called().Get(0).(domain.SergeantMajorConfig)
}

func (m *MockSergeantMajorGame) SetConfig(cfg domain.SergeantMajorConfig) { m.Called(cfg) }

func (m *MockSergeantMajorGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockSergeantMajorGame) GetPhase() domain.SergeantMajorPhase {
	return m.Called().Get(0).(domain.SergeantMajorPhase)
}

func (m *MockSergeantMajorGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockSergeantMajorGame) IsHumanTrumpTurn() bool   { return m.Called().Bool(0) }
func (m *MockSergeantMajorGame) IsHumanDiscardTurn() bool { return m.Called().Bool(0) }
func (m *MockSergeantMajorGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetKittySize() int        { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetDiscardCount() int     { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetLastExchange() int     { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockSergeantMajorGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockSergeantMajorGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockSergeantMajorGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockSergeantMajorGame) GetPlayer(i int) *domain.SergeantMajorPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.SergeantMajorPlayer)
	}
	return nil
}

func (m *MockSergeantMajorGame) GetHint() *domain.SergeantMajorHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.SergeantMajorHint)
	}
	return nil
}

func (m *MockSergeantMajorGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

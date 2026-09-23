//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOmiGame オミゲームモック
type MockOmiGame struct {
	mock.Mock
}

func (m *MockOmiGame) Reset()     { m.Called() }
func (m *MockOmiGame) NextRound() { m.Called() }

func (m *MockOmiGame) PlayerCallTrump(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockOmiGame) CpuCallTrump() { m.Called() }

func (m *MockOmiGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockOmiGame) CpuPlay()      { m.Called() }
func (m *MockOmiGame) ResolveTrick() { m.Called() }
func (m *MockOmiGame) NextTrick()    { m.Called() }
func (m *MockOmiGame) ScoreRound()   { m.Called() }

func (m *MockOmiGame) GetConfig() domain.OmiConfig {
	args := m.Called()
	return args.Get(0).(domain.OmiConfig)
}

func (m *MockOmiGame) SetConfig(cfg domain.OmiConfig) { m.Called(cfg) }

func (m *MockOmiGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOmiGame) GetPhase() domain.OmiPhase {
	args := m.Called()
	return args.Get(0).(domain.OmiPhase)
}

func (m *MockOmiGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOmiGame) IsHumanCallTrumpTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOmiGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOmiGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockOmiGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetTrumpCallerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetCallerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetMakerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockOmiGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOmiGame) GetPlayer(i int) *domain.OmiPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.OmiPlayer)
	}
	return nil
}

func (m *MockOmiGame) GetHint() *domain.OmiHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.OmiHint)
	}
	return nil
}

func (m *MockOmiGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

func (m *MockOmiGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

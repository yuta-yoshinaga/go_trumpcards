//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGaigelGame ガイゲルゲームモック
type MockGaigelGame struct {
	mock.Mock
}

func (m *MockGaigelGame) Reset()     { m.Called() }
func (m *MockGaigelGame) NextRound() { m.Called() }

func (m *MockGaigelGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockGaigelGame) PlayerDeclareMarriage(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockGaigelGame) CpuPlay()      { m.Called() }
func (m *MockGaigelGame) ResolveTrick() { m.Called() }
func (m *MockGaigelGame) NextTrick()    { m.Called() }
func (m *MockGaigelGame) ScoreRound()   { m.Called() }

func (m *MockGaigelGame) GetConfig() domain.GaigelConfig {
	args := m.Called()
	return args.Get(0).(domain.GaigelConfig)
}

func (m *MockGaigelGame) SetConfig(cfg domain.GaigelConfig) { m.Called(cfg) }

func (m *MockGaigelGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockGaigelGame) GetPhase() domain.GaigelPhase {
	args := m.Called()
	return args.Get(0).(domain.GaigelPhase)
}

func (m *MockGaigelGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockGaigelGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockGaigelGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockGaigelGame) GetStockRemaining() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockGaigelGame) GetRoundPoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockGaigelGame) GetRoundMarriagePoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockGaigelGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGaigelGame) GetPlayer(i int) *domain.GaigelPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.GaigelPlayer)
	}
	return nil
}

func (m *MockGaigelGame) GetHint() *domain.GaigelHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.GaigelHint)
	}
	return nil
}

func (m *MockGaigelGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

func (m *MockGaigelGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockGaigelGame) GetMarriageIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

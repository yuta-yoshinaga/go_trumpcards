//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBeloteGame ベロートゲームモック
type MockBeloteGame struct {
	mock.Mock
}

func (m *MockBeloteGame) Reset()     { m.Called() }
func (m *MockBeloteGame) NextRound() { m.Called() }

func (m *MockBeloteGame) PlayerPickUp(orderUp bool) error {
	args := m.Called(orderUp)
	return args.Error(0)
}

func (m *MockBeloteGame) CpuPickUp() { m.Called() }

func (m *MockBeloteGame) PlayerCallTrump(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockBeloteGame) PlayerPassCall() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBeloteGame) CpuCallTrump() { m.Called() }

func (m *MockBeloteGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBeloteGame) CpuPlay()      { m.Called() }
func (m *MockBeloteGame) ResolveTrick() { m.Called() }
func (m *MockBeloteGame) NextTrick()    { m.Called() }
func (m *MockBeloteGame) ScoreRound()   { m.Called() }

func (m *MockBeloteGame) GetConfig() domain.BeloteConfig {
	args := m.Called()
	return args.Get(0).(domain.BeloteConfig)
}

func (m *MockBeloteGame) SetConfig(cfg domain.BeloteConfig) { m.Called(cfg) }

func (m *MockBeloteGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBeloteGame) GetPhase() domain.BelotePhase {
	args := m.Called()
	return args.Get(0).(domain.BelotePhase)
}

func (m *MockBeloteGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBeloteGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBeloteGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockBeloteGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetFaceUpCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockBeloteGame) GetMakerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetMakerPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockBeloteGame) GetRoundPoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockBeloteGame) GetRoundBeloteBonus(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockBeloteGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeloteGame) GetPlayer(i int) *domain.BelotePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.BelotePlayer)
	}
	return nil
}

func (m *MockBeloteGame) GetHint() *domain.BeloteHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.BeloteHint)
	}
	return nil
}

func (m *MockBeloteGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

func (m *MockBeloteGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

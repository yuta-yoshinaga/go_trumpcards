//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCoincheGame コワンシュゲームモック
type MockCoincheGame struct {
	mock.Mock
}

func (m *MockCoincheGame) Reset()     { m.Called() }
func (m *MockCoincheGame) NextRound() { m.Called() }

func (m *MockCoincheGame) PlayerBid(points, suit int) error {
	args := m.Called(points, suit)
	return args.Error(0)
}

func (m *MockCoincheGame) PlayerPassBid() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCoincheGame) CpuBid() { m.Called() }

func (m *MockCoincheGame) PlayerCoinche() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCoincheGame) PlayerSurcoinche() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCoincheGame) PlayerDeclineDouble() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCoincheGame) CpuDouble() { m.Called() }

func (m *MockCoincheGame) GetContractPoints() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetDouble() domain.CoincheDouble {
	args := m.Called()
	return args.Get(0).(domain.CoincheDouble)
}

func (m *MockCoincheGame) GetMultiplier() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetBiddablePoints() []int {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockCoincheGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockCoincheGame) CpuPlay()      { m.Called() }
func (m *MockCoincheGame) ResolveTrick() { m.Called() }
func (m *MockCoincheGame) NextTrick()    { m.Called() }
func (m *MockCoincheGame) ScoreRound()   { m.Called() }

func (m *MockCoincheGame) GetConfig() domain.CoincheConfig {
	args := m.Called()
	return args.Get(0).(domain.CoincheConfig)
}

func (m *MockCoincheGame) SetConfig(cfg domain.CoincheConfig) { m.Called(cfg) }

func (m *MockCoincheGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCoincheGame) GetPhase() domain.CoinchePhase {
	args := m.Called()
	return args.Get(0).(domain.CoinchePhase)
}

func (m *MockCoincheGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCoincheGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCoincheGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockCoincheGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetMakerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetMakerPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockCoincheGame) GetRoundPoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockCoincheGame) GetRoundBeloteBonus(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockCoincheGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCoincheGame) GetPlayer(i int) *domain.CoinchePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.CoinchePlayer)
	}
	return nil
}

func (m *MockCoincheGame) GetHint() *domain.CoincheHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.CoincheHint)
	}
	return nil
}

func (m *MockCoincheGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

func (m *MockCoincheGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

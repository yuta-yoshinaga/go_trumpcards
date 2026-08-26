//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBauernschnapsenGame バウエルンシュナプセンゲームモック
type MockBauernschnapsenGame struct {
	mock.Mock
}

func (m *MockBauernschnapsenGame) Reset()     { m.Called() }
func (m *MockBauernschnapsenGame) NextRound() { m.Called() }

func (m *MockBauernschnapsenGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBauernschnapsenGame) PlayerDeclareMarriage(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBauernschnapsenGame) CpuPlay()      { m.Called() }
func (m *MockBauernschnapsenGame) ResolveTrick() { m.Called() }
func (m *MockBauernschnapsenGame) NextTrick()    { m.Called() }
func (m *MockBauernschnapsenGame) ScoreRound()   { m.Called() }

func (m *MockBauernschnapsenGame) GetConfig() domain.BauernschnapsenConfig {
	args := m.Called()
	return args.Get(0).(domain.BauernschnapsenConfig)
}

func (m *MockBauernschnapsenGame) SetConfig(cfg domain.BauernschnapsenConfig) { m.Called(cfg) }

func (m *MockBauernschnapsenGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBauernschnapsenGame) GetPhase() domain.BauernschnapsenPhase {
	args := m.Called()
	return args.Get(0).(domain.BauernschnapsenPhase)
}

func (m *MockBauernschnapsenGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBauernschnapsenGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockBauernschnapsenGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) DeclareContract(playerIdx int, c domain.BauernschnapsenContract, trumpSuit int) error {
	args := m.Called(playerIdx, c, trumpSuit)
	return args.Error(0)
}

func (m *MockBauernschnapsenGame) CpuDeclareContract() { m.Called() }

func (m *MockBauernschnapsenGame) IsHumanContractTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBauernschnapsenGame) GetContract() domain.BauernschnapsenContract {
	args := m.Called()
	return args.Get(0).(domain.BauernschnapsenContract)
}

func (m *MockBauernschnapsenGame) GetDeclarerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetRoundPoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetRoundMarriagePoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBauernschnapsenGame) GetPlayer(i int) *domain.BauernschnapsenPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.BauernschnapsenPlayer)
	}
	return nil
}

func (m *MockBauernschnapsenGame) GetHint() *domain.BauernschnapsenHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.BauernschnapsenHint)
	}
	return nil
}

func (m *MockBauernschnapsenGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

func (m *MockBauernschnapsenGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockBauernschnapsenGame) GetMarriageIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

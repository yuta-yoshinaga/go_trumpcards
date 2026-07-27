//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockJassGame ヤスゲームモック
type MockJassGame struct {
	mock.Mock
}

func (m *MockJassGame) Reset()     { m.Called() }
func (m *MockJassGame) NextRound() { m.Called() }

func (m *MockJassGame) PlayerChooseTrump(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockJassGame) PlayerSchieben() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockJassGame) CpuBid() { m.Called() }

func (m *MockJassGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockJassGame) CpuPlay()      { m.Called() }
func (m *MockJassGame) ResolveTrick() { m.Called() }
func (m *MockJassGame) NextTrick()    { m.Called() }
func (m *MockJassGame) ScoreRound()   { m.Called() }

func (m *MockJassGame) GetConfig() domain.JassConfig {
	args := m.Called()
	return args.Get(0).(domain.JassConfig)
}

func (m *MockJassGame) SetConfig(cfg domain.JassConfig) { m.Called(cfg) }

func (m *MockJassGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockJassGame) GetPhase() domain.JassPhase {
	args := m.Called()
	return args.Get(0).(domain.JassPhase)
}

func (m *MockJassGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockJassGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockJassGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockJassGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetForehandIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetSchieben() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockJassGame) GetMakerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetMakerPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockJassGame) GetRoundPoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockJassGame) GetRoundWeisPoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockJassGame) GetRoundStockPoints(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockJassGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockJassGame) GetPlayer(i int) *domain.JassPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.JassPlayer)
	}
	return nil
}

func (m *MockJassGame) GetHint() *domain.JassHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.JassHint)
	}
	return nil
}

func (m *MockJassGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

func (m *MockJassGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

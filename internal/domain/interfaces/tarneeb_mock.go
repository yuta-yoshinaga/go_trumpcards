//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTarneebGame Tarneeb ゲームモック
type MockTarneebGame struct {
	mock.Mock
}

func (m *MockTarneebGame) Reset() { m.Called() }

func (m *MockTarneebGame) NextRound() { m.Called() }

func (m *MockTarneebGame) PlayerBid(bid int) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockTarneebGame) CpuBid() { m.Called() }

func (m *MockTarneebGame) PlayerDeclareTrump(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockTarneebGame) CpuDeclareTrump() { m.Called() }

func (m *MockTarneebGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockTarneebGame) CpuPlay() { m.Called() }

func (m *MockTarneebGame) ResolveTrick() { m.Called() }

func (m *MockTarneebGame) NextTrick() { m.Called() }

func (m *MockTarneebGame) ScoreRound() { m.Called() }

func (m *MockTarneebGame) GetConfig() domain.TarneebConfig {
	args := m.Called()
	return args.Get(0).(domain.TarneebConfig)
}

func (m *MockTarneebGame) SetConfig(cfg domain.TarneebConfig) { m.Called(cfg) }

func (m *MockTarneebGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTarneebGame) GetPhase() domain.TarneebPhase {
	args := m.Called()
	return args.Get(0).(domain.TarneebPhase)
}

func (m *MockTarneebGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTarneebGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTarneebGame) IsHumanTrumpTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTarneebGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockTarneebGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetBidWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetHighestBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetRedealCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockTarneebGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTarneebGame) GetPlayer(i int) *domain.TarneebPlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.TarneebPlayer)
}

func (m *MockTarneebGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	return args.Get(0).([]int)
}

// GetHint モック
func (m *MockTarneebGame) GetHint() *domain.TarneebHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.TarneebHint); ok {
		return val
	}
	return nil
}

func (m *MockTarneebGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}

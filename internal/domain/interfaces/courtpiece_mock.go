//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCourtPieceGame Court Piece ゲームモック
type MockCourtPieceGame struct {
	mock.Mock
}

func (m *MockCourtPieceGame) Reset() { m.Called() }

func (m *MockCourtPieceGame) NextRound() { m.Called() }

func (m *MockCourtPieceGame) PlayerDeclareTrump(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockCourtPieceGame) CpuDeclareTrump() { m.Called() }

func (m *MockCourtPieceGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockCourtPieceGame) CpuPlay() { m.Called() }

func (m *MockCourtPieceGame) ResolveTrick() { m.Called() }

func (m *MockCourtPieceGame) NextTrick() { m.Called() }

func (m *MockCourtPieceGame) ScoreRound() { m.Called() }

func (m *MockCourtPieceGame) GetConfig() domain.CourtPieceConfig {
	args := m.Called()
	return args.Get(0).(domain.CourtPieceConfig)
}

func (m *MockCourtPieceGame) SetConfig(cfg domain.CourtPieceConfig) { m.Called(cfg) }

func (m *MockCourtPieceGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCourtPieceGame) GetPhase() domain.CourtPiecePhase {
	args := m.Called()
	return args.Get(0).(domain.CourtPiecePhase)
}

func (m *MockCourtPieceGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCourtPieceGame) IsHumanTrumpTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCourtPieceGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockCourtPieceGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetCallerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetConsecutiveWins() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetLastWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) IsLastRoundCourt() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCourtPieceGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCourtPieceGame) GetPlayer(i int) *domain.CourtPiecePlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.CourtPiecePlayer)
}

func (m *MockCourtPieceGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	return args.Get(0).([]int)
}

// GetHint モック
func (m *MockCourtPieceGame) GetHint() *domain.CourtPieceHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.CourtPieceHint); ok {
		return val
	}
	return nil
}

func (m *MockCourtPieceGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}

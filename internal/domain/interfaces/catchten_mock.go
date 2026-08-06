//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCatchTenGame Catch the Ten ゲームモック
type MockCatchTenGame struct {
	mock.Mock
}

func (m *MockCatchTenGame) Reset() {
	m.Called()
}

func (m *MockCatchTenGame) NextRound() {
	m.Called()
}

func (m *MockCatchTenGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockCatchTenGame) CpuPlay() {
	m.Called()
}

func (m *MockCatchTenGame) ResolveTrick() {
	m.Called()
}

func (m *MockCatchTenGame) NextTrick() {
	m.Called()
}

func (m *MockCatchTenGame) ScoreRound() {
	m.Called()
}

func (m *MockCatchTenGame) GetConfig() domain.CatchTenConfig {
	args := m.Called()
	return args.Get(0).(domain.CatchTenConfig)
}

func (m *MockCatchTenGame) SetConfig(cfg domain.CatchTenConfig) {
	m.Called(cfg)
}

func (m *MockCatchTenGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCatchTenGame) GetPhase() domain.CatchTenPhase {
	args := m.Called()
	return args.Get(0).(domain.CatchTenPhase)
}

func (m *MockCatchTenGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockCatchTenGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCatchTenGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCatchTenGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCatchTenGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockCatchTenGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCatchTenGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCatchTenGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCatchTenGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockCatchTenGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCatchTenGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCatchTenGame) GetPlayer(i int) *domain.CatchTenPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.CatchTenPlayer)
	}
	return nil
}

func (m *MockCatchTenGame) GetHint() *domain.CatchTenHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.CatchTenHint)
	}
	return nil
}

// GetValidPlayIndices はプレイ可能なカードのインデックスを返すモック。
func (m *MockCatchTenGame) GetValidPlayIndices(playerIdx int) []int {
	out, _ := m.Called(playerIdx).Get(0).([]int)
	return out
}

func (m *MockCatchTenGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

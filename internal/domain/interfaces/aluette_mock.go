//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAluetteGame アリュエットのゲームモック
type MockAluetteGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockAluetteGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockAluetteGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockAluetteGame) PlayerPlay(cardIndex int) error {
	args := _m.Called(cardIndex)
	return args.Error(0)
}

// CpuPlay モック
func (_m *MockAluetteGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockAluetteGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockAluetteGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockAluetteGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockAluetteGame) GetConfig() domain.AluetteConfig {
	args := _m.Called()
	return args.Get(0).(domain.AluetteConfig)
}

// SetConfig モック
func (_m *MockAluetteGame) SetConfig(cfg domain.AluetteConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockAluetteGame) GetGameEndFlag() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetPhase モック
func (_m *MockAluetteGame) GetPhase() domain.AluettePhase {
	args := _m.Called()
	return args.Get(0).(domain.AluettePhase)
}

// IsHumanTurn モック
func (_m *MockAluetteGame) IsHumanTurn() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetRoundNumber モック
func (_m *MockAluetteGame) GetRoundNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTrickNumber モック
func (_m *MockAluetteGame) GetTrickNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockAluetteGame) GetCurrentPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentTrick モック
func (_m *MockAluetteGame) GetCurrentTrick() []*domain.TrickCard {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockAluetteGame) GetLeadPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetDealerIdx モック
func (_m *MockAluetteGame) GetDealerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetLastTrickWinner モック
func (_m *MockAluetteGame) GetLastTrickWinner() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTeamScores モック
func (_m *MockAluetteGame) GetTeamScores() [2]int {
	args := _m.Called()
	return args.Get(0).([2]int)
}

// GetRoundTricks モック
func (_m *MockAluetteGame) GetRoundTricks() [domain.AluettePlayerCnt]int {
	args := _m.Called()
	return args.Get(0).([domain.AluettePlayerCnt]int)
}

// GetWinnerTeam モック
func (_m *MockAluetteGame) GetWinnerTeam() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayerCnt モック
func (_m *MockAluetteGame) GetPlayerCnt() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayer モック
func (_m *MockAluetteGame) GetPlayer(i int) *domain.AluettePlayer {
	args := _m.Called(i)
	if v, ok := args.Get(0).(*domain.AluettePlayer); ok {
		return v
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockAluetteGame) GetPlayableIndices(playerIdx int) []int {
	args := _m.Called(playerIdx)
	if v, ok := args.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockAluetteGame) GetHint() *domain.AluetteHint {
	args := _m.Called()
	if v, ok := args.Get(0).(*domain.AluetteHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockAluetteGame) GetActionLog() []*domain.ActionLogEntry {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

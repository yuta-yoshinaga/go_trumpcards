//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMinchiateGame ミンキアーテのゲームモック
type MockMinchiateGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockMinchiateGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockMinchiateGame) NextRound() {
	_m.Called()
}

// PlayerScarto モック
func (_m *MockMinchiateGame) PlayerScarto(cardIndices []int) error {
	args := _m.Called(cardIndices)
	return args.Error(0)
}

// CpuScarto モック
func (_m *MockMinchiateGame) CpuScarto() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockMinchiateGame) PlayerPlay(cardIndex int) error {
	args := _m.Called(cardIndex)
	return args.Error(0)
}

// CpuPlay モック
func (_m *MockMinchiateGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockMinchiateGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockMinchiateGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockMinchiateGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockMinchiateGame) GetConfig() domain.MinchiateConfig {
	args := _m.Called()
	return args.Get(0).(domain.MinchiateConfig)
}

// SetConfig モック
func (_m *MockMinchiateGame) SetConfig(cfg domain.MinchiateConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockMinchiateGame) GetGameEndFlag() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetPhase モック
func (_m *MockMinchiateGame) GetPhase() domain.MinchiatePhase {
	args := _m.Called()
	return args.Get(0).(domain.MinchiatePhase)
}

// IsHumanTurn モック
func (_m *MockMinchiateGame) IsHumanTurn() bool {
	args := _m.Called()
	return args.Bool(0)
}

// IsHumanScartoTurn モック
func (_m *MockMinchiateGame) IsHumanScartoTurn() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetRoundNumber モック
func (_m *MockMinchiateGame) GetRoundNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTrickNumber モック
func (_m *MockMinchiateGame) GetTrickNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockMinchiateGame) GetCurrentPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentTrick モック
func (_m *MockMinchiateGame) GetCurrentTrick() []*domain.TrickCard {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockMinchiateGame) GetLeadPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetDealerIdx モック
func (_m *MockMinchiateGame) GetDealerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetLastTrickWinner モック
func (_m *MockMinchiateGame) GetLastTrickWinner() int {
	args := _m.Called()
	return args.Int(0)
}

// GetScartoSize モック
func (_m *MockMinchiateGame) GetScartoSize() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTeamScores モック
func (_m *MockMinchiateGame) GetTeamScores() [2]int {
	args := _m.Called()
	return args.Get(0).([2]int)
}

// GetRoundTricks モック
func (_m *MockMinchiateGame) GetRoundTricks() [domain.MinchiatePlayerCnt]int {
	args := _m.Called()
	return args.Get(0).([domain.MinchiatePlayerCnt]int)
}

// GetWinnerTeam モック
func (_m *MockMinchiateGame) GetWinnerTeam() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayerCnt モック
func (_m *MockMinchiateGame) GetPlayerCnt() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayer モック
func (_m *MockMinchiateGame) GetPlayer(i int) *domain.MinchiatePlayer {
	args := _m.Called(i)
	if v, ok := args.Get(0).(*domain.MinchiatePlayer); ok {
		return v
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockMinchiateGame) GetPlayableIndices(playerIdx int) []int {
	args := _m.Called(playerIdx)
	if v, ok := args.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockMinchiateGame) GetHint() *domain.MinchiateHint {
	args := _m.Called()
	if v, ok := args.Get(0).(*domain.MinchiateHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockMinchiateGame) GetActionLog() []*domain.ActionLogEntry {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

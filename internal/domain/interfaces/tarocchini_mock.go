//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTarocchiniGame タロッキーニのゲームモック
type MockTarocchiniGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTarocchiniGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockTarocchiniGame) NextRound() {
	_m.Called()
}

// PlayerScarto モック
func (_m *MockTarocchiniGame) PlayerScarto(cardIndices []int) error {
	args := _m.Called(cardIndices)
	return args.Error(0)
}

// CpuScarto モック
func (_m *MockTarocchiniGame) CpuScarto() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockTarocchiniGame) PlayerPlay(cardIndex int) error {
	args := _m.Called(cardIndex)
	return args.Error(0)
}

// CpuPlay モック
func (_m *MockTarocchiniGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockTarocchiniGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockTarocchiniGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockTarocchiniGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockTarocchiniGame) GetConfig() domain.TarocchiniConfig {
	args := _m.Called()
	return args.Get(0).(domain.TarocchiniConfig)
}

// SetConfig モック
func (_m *MockTarocchiniGame) SetConfig(cfg domain.TarocchiniConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockTarocchiniGame) GetGameEndFlag() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetPhase モック
func (_m *MockTarocchiniGame) GetPhase() domain.TarocchiniPhase {
	args := _m.Called()
	return args.Get(0).(domain.TarocchiniPhase)
}

// IsHumanTurn モック
func (_m *MockTarocchiniGame) IsHumanTurn() bool {
	args := _m.Called()
	return args.Bool(0)
}

// IsHumanScartoTurn モック
func (_m *MockTarocchiniGame) IsHumanScartoTurn() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetRoundNumber モック
func (_m *MockTarocchiniGame) GetRoundNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTrickNumber モック
func (_m *MockTarocchiniGame) GetTrickNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockTarocchiniGame) GetCurrentPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentTrick モック
func (_m *MockTarocchiniGame) GetCurrentTrick() []*domain.TrickCard {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockTarocchiniGame) GetLeadPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetDealerIdx モック
func (_m *MockTarocchiniGame) GetDealerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetLastTrickWinner モック
func (_m *MockTarocchiniGame) GetLastTrickWinner() int {
	args := _m.Called()
	return args.Int(0)
}

// GetScartoSize モック
func (_m *MockTarocchiniGame) GetScartoSize() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTeamScores モック
func (_m *MockTarocchiniGame) GetTeamScores() [2]int {
	args := _m.Called()
	return args.Get(0).([2]int)
}

// GetRoundTricks モック
func (_m *MockTarocchiniGame) GetRoundTricks() [domain.TarocchiniPlayerCnt]int {
	args := _m.Called()
	return args.Get(0).([domain.TarocchiniPlayerCnt]int)
}

// GetWinnerTeam モック
func (_m *MockTarocchiniGame) GetWinnerTeam() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayerCnt モック
func (_m *MockTarocchiniGame) GetPlayerCnt() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayer モック
func (_m *MockTarocchiniGame) GetPlayer(i int) *domain.TarocchiniPlayer {
	args := _m.Called(i)
	if v, ok := args.Get(0).(*domain.TarocchiniPlayer); ok {
		return v
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockTarocchiniGame) GetPlayableIndices(playerIdx int) []int {
	args := _m.Called(playerIdx)
	if v, ok := args.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockTarocchiniGame) GetHint() *domain.TarocchiniHint {
	args := _m.Called()
	if v, ok := args.Get(0).(*domain.TarocchiniHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockTarocchiniGame) GetActionLog() []*domain.ActionLogEntry {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

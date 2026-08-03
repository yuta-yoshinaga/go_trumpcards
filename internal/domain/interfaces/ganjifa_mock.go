//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGanjifaGame ガンジファのゲームモック
type MockGanjifaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockGanjifaGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockGanjifaGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockGanjifaGame) PlayerPlay(cardIndex int) error {
	args := _m.Called(cardIndex)
	return args.Error(0)
}

// CpuPlay モック
func (_m *MockGanjifaGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockGanjifaGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockGanjifaGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockGanjifaGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockGanjifaGame) GetConfig() domain.GanjifaConfig {
	args := _m.Called()
	return args.Get(0).(domain.GanjifaConfig)
}

// SetConfig モック
func (_m *MockGanjifaGame) SetConfig(cfg domain.GanjifaConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockGanjifaGame) GetGameEndFlag() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetPhase モック
func (_m *MockGanjifaGame) GetPhase() domain.GanjifaPhase {
	args := _m.Called()
	return args.Get(0).(domain.GanjifaPhase)
}

// IsHumanTurn モック
func (_m *MockGanjifaGame) IsHumanTurn() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetRoundNumber モック
func (_m *MockGanjifaGame) GetRoundNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTrickNumber モック
func (_m *MockGanjifaGame) GetTrickNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockGanjifaGame) GetCurrentPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentTrick モック
func (_m *MockGanjifaGame) GetCurrentTrick() []*domain.TrickCard {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockGanjifaGame) GetLeadPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetDealerIdx モック
func (_m *MockGanjifaGame) GetDealerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTrumpSuit モック
func (_m *MockGanjifaGame) GetTrumpSuit() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayerScores モック
func (_m *MockGanjifaGame) GetPlayerScores() [domain.GanjifaPlayerCnt]int {
	args := _m.Called()
	return args.Get(0).([domain.GanjifaPlayerCnt]int)
}

// GetRoundTricks モック
func (_m *MockGanjifaGame) GetRoundTricks() [domain.GanjifaPlayerCnt]int {
	args := _m.Called()
	return args.Get(0).([domain.GanjifaPlayerCnt]int)
}

// GetWinnerPlayer モック
func (_m *MockGanjifaGame) GetWinnerPlayer() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayerCnt モック
func (_m *MockGanjifaGame) GetPlayerCnt() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayer モック
func (_m *MockGanjifaGame) GetPlayer(i int) *domain.GanjifaPlayer {
	args := _m.Called(i)
	if v, ok := args.Get(0).(*domain.GanjifaPlayer); ok {
		return v
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockGanjifaGame) GetPlayableIndices(playerIdx int) []int {
	args := _m.Called(playerIdx)
	if v, ok := args.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockGanjifaGame) GetHint() *domain.GanjifaHint {
	args := _m.Called()
	if v, ok := args.Get(0).(*domain.GanjifaHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockGanjifaGame) GetActionLog() []*domain.ActionLogEntry {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

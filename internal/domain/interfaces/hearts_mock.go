//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHeartsGame ハーツゲームモック
type MockHeartsGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockHeartsGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockHeartsGame) NextRound() {
	_m.Called()
}

// PlayerPass モック
func (_m *MockHeartsGame) PlayerPass(cardIndices []int) error {
	ret := _m.Called(cardIndices)
	return ret.Error(0)
}

// CpuPass モック
func (_m *MockHeartsGame) CpuPass() {
	_m.Called()
}

// ExecutePass モック
func (_m *MockHeartsGame) ExecutePass() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockHeartsGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockHeartsGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockHeartsGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockHeartsGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockHeartsGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockHeartsGame) GetConfig() domain.HeartsConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.HeartsConfig)
}

// SetConfig モック
func (_m *MockHeartsGame) SetConfig(cfg domain.HeartsConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockHeartsGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockHeartsGame) GetPhase() domain.HeartsPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.HeartsPhase)
}

// IsHumanTurn モック
func (_m *MockHeartsGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRoundNumber モック
func (_m *MockHeartsGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrickNumber モック
func (_m *MockHeartsGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockHeartsGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTrick モック
func (_m *MockHeartsGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return val
	}
	return nil
}

// GetHeartsBroken モック
func (_m *MockHeartsGame) GetHeartsBroken() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPassDirection モック
func (_m *MockHeartsGame) GetPassDirection() domain.HeartsPassDirection {
	ret := _m.Called()
	return ret.Get(0).(domain.HeartsPassDirection)
}

// GetLeadPlayerIdx モック
func (_m *MockHeartsGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetWinnerIdx モック
func (_m *MockHeartsGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayerCnt モック
func (_m *MockHeartsGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockHeartsGame) GetPlayer(i int) *domain.HeartsPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.HeartsPlayer); ok {
		return val
	}
	return nil
}

// GetPassReady モック
func (_m *MockHeartsGame) GetPassReady() [domain.HeartsPlayerCnt]bool {
	ret := _m.Called()
	return ret.Get(0).([domain.HeartsPlayerCnt]bool)
}

// GetPassedCards モック
func (_m *MockHeartsGame) GetPassedCards() [domain.HeartsPlayerCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.HeartsPlayerCnt][]*domain.Card)
}

// GetHint モック
func (_m *MockHeartsGame) GetHint() *domain.HeartsHint {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.HeartsHint); ok {
		return val
	}
	return nil
}

// GetActionLog モック
func (_m *MockHeartsGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

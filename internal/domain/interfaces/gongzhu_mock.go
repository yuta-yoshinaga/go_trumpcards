//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGongZhuGame 拱猪ゲームモック
type MockGongZhuGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockGongZhuGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockGongZhuGame) NextRound() {
	_m.Called()
}

// PlayerExpose モック
func (_m *MockGongZhuGame) PlayerExpose(cardIndices []int) error {
	ret := _m.Called(cardIndices)
	return ret.Error(0)
}

// CpuExpose モック
func (_m *MockGongZhuGame) CpuExpose() {
	_m.Called()
}

// ExecuteExpose モック
func (_m *MockGongZhuGame) ExecuteExpose() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockGongZhuGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockGongZhuGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockGongZhuGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockGongZhuGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockGongZhuGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockGongZhuGame) GetConfig() domain.GongZhuConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GongZhuConfig)
}

// SetConfig モック
func (_m *MockGongZhuGame) SetConfig(cfg domain.GongZhuConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockGongZhuGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockGongZhuGame) GetPhase() domain.GongZhuPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GongZhuPhase)
}

// IsHumanTurn モック
func (_m *MockGongZhuGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRoundNumber モック
func (_m *MockGongZhuGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrickNumber モック
func (_m *MockGongZhuGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockGongZhuGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayableIndices モック
func (_m *MockGongZhuGame) GetPlayableIndices(playerIdx int) []int {
	out, _ := _m.Called(playerIdx).Get(0).([]int)
	return out
}

// GetCurrentTrick モック
func (_m *MockGongZhuGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return val
	}
	return nil
}

// GetHeartsBroken モック
func (_m *MockGongZhuGame) GetHeartsBroken() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetExposure モック
func (_m *MockGongZhuGame) GetExposure() domain.GongZhuExposure {
	ret := _m.Called()
	return ret.Get(0).(domain.GongZhuExposure)
}

// GetExposeReady モック
func (_m *MockGongZhuGame) GetExposeReady() [domain.GongZhuPlayerCnt]bool {
	ret := _m.Called()
	return ret.Get(0).([domain.GongZhuPlayerCnt]bool)
}

// GetExposableIndices モック
func (_m *MockGongZhuGame) GetExposableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockGongZhuGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetWinnerIdx モック
func (_m *MockGongZhuGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayerCnt モック
func (_m *MockGongZhuGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockGongZhuGame) GetPlayer(i int) *domain.GongZhuPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.GongZhuPlayer); ok {
		return val
	}
	return nil
}

// GetHint モック
func (_m *MockGongZhuGame) GetHint() *domain.GongZhuHint {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.GongZhuHint); ok {
		return val
	}
	return nil
}

// GetActionLog モック
func (_m *MockGongZhuGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

// ScoreBreakdownFor モック
func (_m *MockGongZhuGame) ScoreBreakdownFor(playerIdx int) domain.GongZhuScoreBreakdown {
	ret := _m.Called(playerIdx)
	if val, ok := ret.Get(0).(domain.GongZhuScoreBreakdown); ok {
		return val
	}
	return domain.GongZhuScoreBreakdown{}
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpeedGame スピードゲームモック
type MockSpeedGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSpeedGame) Reset() { _m.Called() }

// PlayerPlay モック
func (_m *MockSpeedGame) PlayerPlay(cardIndex, pileIndex int) error {
	ret := _m.Called(cardIndex, pileIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSpeedGame) CpuPlay() []*domain.SpeedCpuAction {
	ret := _m.Called()
	return ret.Get(0).([]*domain.SpeedCpuAction)
}

// Flip モック
func (_m *MockSpeedGame) Flip() error {
	ret := _m.Called()
	return ret.Error(0)
}

// UpdatePhase モック
func (_m *MockSpeedGame) UpdatePhase() { _m.Called() }

// GetHint モック
func (_m *MockSpeedGame) GetHint() (int, int, bool) {
	ret := _m.Called()
	return ret.Int(0), ret.Int(1), ret.Bool(2)
}

// GetConfig モック
func (_m *MockSpeedGame) GetConfig() domain.SpeedConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SpeedConfig)
}

// SetConfig モック
func (_m *MockSpeedGame) SetConfig(cfg domain.SpeedConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockSpeedGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockSpeedGame) GetPhase() domain.SpeedPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SpeedPhase)
}

// IsHumanTurn モック
func (_m *MockSpeedGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsStuck モック
func (_m *MockSpeedGame) IsStuck() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerCnt モック
func (_m *MockSpeedGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockSpeedGame) GetPlayer(i int) *domain.SpeedPlayer {
	ret := _m.Called(i)
	return ret.Get(0).(*domain.SpeedPlayer)
}

// GetCenterPile モック
func (_m *MockSpeedGame) GetCenterPile(i int) *domain.Card {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockSpeedGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// CanPlay モック
func (_m *MockSpeedGame) CanPlay(playerIdx, cardIdx, pileIdx int) bool {
	ret := _m.Called(playerIdx, cardIdx, pileIdx)
	return ret.Bool(0)
}

// PlayerHasAnyPlay モック
func (_m *MockSpeedGame) PlayerHasAnyPlay(playerIdx int) bool {
	ret := _m.Called(playerIdx)
	return ret.Bool(0)
}

// GetActionLog モック
func (_m *MockSpeedGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	return ret.Get(0).([]*domain.ActionLogEntry)
}

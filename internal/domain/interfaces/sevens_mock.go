//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSevensGame 7並べゲームモック
type MockSevensGame struct {
	mock.Mock
}

// SetConfig モック
func (_m *MockSevensGame) SetConfig(config domain.SevensConfig) {
	_m.Called(config)
}

// Reset モック
func (_m *MockSevensGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockSevensGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockSevensGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// HasAnyOption モック
func (_m *MockSevensGame) HasAnyOption(playerIdx int) bool {
	ret := _m.Called(playerIdx)
	return ret.Bool(0)
}

// AutoHandleNoOption モック
func (_m *MockSevensGame) AutoHandleNoOption() {
	_m.Called()
}

// CpuPlay モック
func (_m *MockSevensGame) CpuPlay() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockSevensGame) PlayerPlay(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

// PlayerPlayJoker モック
func (_m *MockSevensGame) PlayerPlayJoker(cardIdx, targetSuit, targetValue int) error {
	ret := _m.Called(cardIdx, targetSuit, targetValue)
	return ret.Error(0)
}

// GetCurrentTurn モック
func (_m *MockSevensGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayerCnt モック
func (_m *MockSevensGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockSevensGame) GetPlayer(i int) *domain.SevensPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.SevensPlayer); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockSevensGame) GetConfig() domain.SevensConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SevensConfig)
}

// GetTableMinVals モック
func (_m *MockSevensGame) GetTableMinVals() [5]int {
	ret := _m.Called()
	return ret.Get(0).([5]int)
}

// GetTableMaxVals モック
func (_m *MockSevensGame) GetTableMaxVals() [5]int {
	ret := _m.Called()
	return ret.Get(0).([5]int)
}

// GetHumanAction モック
func (_m *MockSevensGame) GetHumanAction() *domain.SevensCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.SevensCpuAction); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockSevensGame) GetCpuActions() []*domain.SevensCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.SevensCpuAction); ok {
		return val
	}
	return nil
}

// GetTablePlaced モック
func (_m *MockSevensGame) GetTablePlaced() [5]uint16 {
	ret := _m.Called()
	return ret.Get(0).([5]uint16)
}

// GetPlayableCardIndices モック
func (_m *MockSevensGame) GetPlayableCardIndices() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

// GetActionLog モック
func (_m *MockSevensGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

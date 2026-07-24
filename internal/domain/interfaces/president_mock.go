//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPresidentGame プレジデントゲームモック
type MockPresidentGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPresidentGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockPresidentGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockPresidentGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockPresidentGame) PlayerPlay(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// SuggestWeakestPlay モック
func (_m *MockPresidentGame) SuggestWeakestPlay(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// CpuPlay モック
func (_m *MockPresidentGame) CpuPlay() {
	_m.Called()
}

// SetConfig モック
func (_m *MockPresidentGame) SetConfig(config domain.PresidentConfig) {
	_m.Called(config)
}

// GetPlayerCnt モック
func (_m *MockPresidentGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockPresidentGame) GetPlayer(i int) *domain.PresidentPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.PresidentPlayer); ok {
		return val
	}
	return nil
}

// GetRevolutionActive モック
func (_m *MockPresidentGame) GetRevolutionActive() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetExchangeActions モック
func (_m *MockPresidentGame) GetExchangeActions() []*domain.PresidentExchangeAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.PresidentExchangeAction); ok {
		return val
	}
	return nil
}

// GetTableCards モック
func (_m *MockPresidentGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetLastPlayPlayerIdx モック
func (_m *MockPresidentGame) GetLastPlayPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanAction モック
func (_m *MockPresidentGame) GetHumanAction() *domain.PresidentCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.PresidentCpuAction); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockPresidentGame) GetCpuActions() []*domain.PresidentCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.PresidentCpuAction); ok {
		return val
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockPresidentGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockPresidentGame) GetConfig() domain.PresidentConfig {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.PresidentConfig); ok {
		return val
	}
	return domain.PresidentConfig{}
}

// GetPassCount モック
func (_m *MockPresidentGame) GetPassCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockPresidentGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

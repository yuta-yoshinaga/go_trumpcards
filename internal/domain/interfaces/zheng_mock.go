//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockZhengGame 争上游ゲームモック
type MockZhengGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockZhengGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockZhengGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockZhengGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockZhengGame) PlayerPlay(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockZhengGame) CpuPlay() {
	_m.Called()
}

// HasPendingAction モック
func (_m *MockZhengGame) HasPendingAction() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// SetConfig モック
func (_m *MockZhengGame) SetConfig(config domain.ZhengConfig) {
	_m.Called(config)
}

// GetPlayerCnt モック
func (_m *MockZhengGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockZhengGame) GetPlayer(i int) *domain.ZhengPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.ZhengPlayer); ok {
		return v
	}
	return nil
}

// GetTableCards モック
func (_m *MockZhengGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetTablePlayType モック
func (_m *MockZhengGame) GetTablePlayType() domain.ZhengPlayType {
	ret := _m.Called()
	return ret.Get(0).(domain.ZhengPlayType)
}

// GetLastPlayPlayerIdx モック
func (_m *MockZhengGame) GetLastPlayPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanAction モック
func (_m *MockZhengGame) GetHumanAction() *domain.ZhengAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.ZhengAction); ok {
		return v
	}
	return nil
}

// GetCpuActions モック
func (_m *MockZhengGame) GetCpuActions() []*domain.ZhengAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ZhengAction); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockZhengGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockZhengGame) GetConfig() domain.ZhengConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ZhengConfig)
}

// GetPassCount モック
func (_m *MockZhengGame) GetPassCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockZhengGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

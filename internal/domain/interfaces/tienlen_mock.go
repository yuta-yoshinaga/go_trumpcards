//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTienLenGame Tien Lenゲームモック
type MockTienLenGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTienLenGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockTienLenGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockTienLenGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockTienLenGame) PlayerPlay(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockTienLenGame) CpuPlay() {
	_m.Called()
}

// HasPendingAction モック
func (_m *MockTienLenGame) HasPendingAction() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// SetConfig モック
func (_m *MockTienLenGame) SetConfig(config domain.TienLenConfig) {
	_m.Called(config)
}

// GetPlayerCnt モック
func (_m *MockTienLenGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockTienLenGame) GetPlayer(i int) *domain.TienLenPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.TienLenPlayer); ok {
		return v
	}
	return nil
}

// GetTableCards モック
func (_m *MockTienLenGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetTablePlayType モック
func (_m *MockTienLenGame) GetTablePlayType() domain.TienLenPlayType {
	ret := _m.Called()
	return ret.Get(0).(domain.TienLenPlayType)
}

// GetLastPlayPlayerIdx モック
func (_m *MockTienLenGame) GetLastPlayPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanAction モック
func (_m *MockTienLenGame) GetHumanAction() *domain.TienLenAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.TienLenAction); ok {
		return v
	}
	return nil
}

// GetCpuActions モック
func (_m *MockTienLenGame) GetCpuActions() []*domain.TienLenAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TienLenAction); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockTienLenGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockTienLenGame) GetConfig() domain.TienLenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TienLenConfig)
}

// GetPassCount モック
func (_m *MockTienLenGame) GetPassCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockTienLenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockTienLenGame) GetHint() *domain.TienLenHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.TienLenHint)
}

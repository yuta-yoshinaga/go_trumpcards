//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBigTwoGame Big Twoゲームモック
type MockBigTwoGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockBigTwoGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockBigTwoGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockBigTwoGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockBigTwoGame) PlayerPlay(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockBigTwoGame) CpuPlay() {
	_m.Called()
}

// HasPendingAction モック
func (_m *MockBigTwoGame) HasPendingAction() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// SetConfig モック
func (_m *MockBigTwoGame) SetConfig(config domain.BigTwoConfig) {
	_m.Called(config)
}

// GetPlayerCnt モック
func (_m *MockBigTwoGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockBigTwoGame) GetPlayer(i int) *domain.BigTwoPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.BigTwoPlayer); ok {
		return v
	}
	return nil
}

// GetTableCards モック
func (_m *MockBigTwoGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetTablePlayType モック
func (_m *MockBigTwoGame) GetTablePlayType() domain.BigTwoPlayType {
	ret := _m.Called()
	return ret.Get(0).(domain.BigTwoPlayType)
}

// GetLastPlayPlayerIdx モック
func (_m *MockBigTwoGame) GetLastPlayPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanAction モック
func (_m *MockBigTwoGame) GetHumanAction() *domain.BigTwoAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.BigTwoAction); ok {
		return v
	}
	return nil
}

// GetCpuActions モック
func (_m *MockBigTwoGame) GetCpuActions() []*domain.BigTwoAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.BigTwoAction); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockBigTwoGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockBigTwoGame) GetConfig() domain.BigTwoConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BigTwoConfig)
}

// GetPassCount モック
func (_m *MockBigTwoGame) GetPassCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockBigTwoGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

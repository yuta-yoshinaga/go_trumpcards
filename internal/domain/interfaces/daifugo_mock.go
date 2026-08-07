//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDaifugoGame 大富豪ゲームモック
type MockDaifugoGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockDaifugoGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockDaifugoGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockDaifugoGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockDaifugoGame) PlayerPlay(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockDaifugoGame) CpuPlay() {
	_m.Called()
}

// GetPlayerCnt モック
func (_m *MockDaifugoGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockDaifugoGame) GetPlayer(i int) *domain.DaifugoPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.DaifugoPlayer); ok {
		return val
	}
	return nil
}

// GetRevolutionActive モック
func (_m *MockDaifugoGame) GetRevolutionActive() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetElevenBackActive モック
func (_m *MockDaifugoGame) GetElevenBackActive() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetSuitLocked モック
func (_m *MockDaifugoGame) GetSuitLocked() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLockedSuit モック
func (_m *MockDaifugoGame) GetLockedSuit() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTableIsSequence モック
func (_m *MockDaifugoGame) GetTableIsSequence() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetExchangeActions モック
func (_m *MockDaifugoGame) GetExchangeActions() []*domain.DaifugoExchangeAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.DaifugoExchangeAction); ok {
		return val
	}
	return nil
}

// GetTableCards モック
func (_m *MockDaifugoGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetLastPlayPlayerIdx モック
func (_m *MockDaifugoGame) GetLastPlayPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanAction モック
func (_m *MockDaifugoGame) GetHumanAction() *domain.DaifugoCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.DaifugoCpuAction); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockDaifugoGame) GetCpuActions() []*domain.DaifugoCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.DaifugoCpuAction); ok {
		return val
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockDaifugoGame) GetPlayableCardIndices() []int {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).([]int)
}

func (_m *MockDaifugoGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockDaifugoGame) GetConfig() domain.DaifugoConfig {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.DaifugoConfig); ok {
		return val
	}
	return domain.DaifugoConfig{}
}

// GetPassCount モック
func (_m *MockDaifugoGame) GetPassCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// HasPendingAction モック
func (_m *MockDaifugoGame) HasPendingAction() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// SetConfig モック
func (_m *MockDaifugoGame) SetConfig(config domain.DaifugoConfig) {
	_m.Called(config)
}

// GetPendingActionType モック
func (_m *MockDaifugoGame) GetPendingActionType() domain.DaifugoPendingAction {
	ret := _m.Called()
	return ret.Get(0).(domain.DaifugoPendingAction)
}

// GetPendingActionTarget モック
func (_m *MockDaifugoGame) GetPendingActionTarget() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetReverseDirection モック
func (_m *MockDaifugoGame) GetReverseDirection() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetNumberLocked モック
func (_m *MockDaifugoGame) GetNumberLocked() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetSequenceLocked モック
func (_m *MockDaifugoGame) GetSequenceLocked() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// SortHumanHand モック
func (_m *MockDaifugoGame) SortHumanHand(mode domain.DaifugoSortMode) error {
	ret := _m.Called(mode)
	return ret.Error(0)
}

// GetSortMode モック
func (_m *MockDaifugoGame) GetSortMode() domain.DaifugoSortMode {
	ret := _m.Called()
	return ret.Get(0).(domain.DaifugoSortMode)
}

// GetActionLog モック
func (_m *MockDaifugoGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

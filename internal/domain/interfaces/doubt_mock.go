package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDoubtGame ダウトゲームモック
type MockDoubtGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockDoubtGame) Reset() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockDoubtGame) PlayerPlay(cardIndices []int, claimedValue int) error {
	ret := _m.Called(cardIndices, claimedValue)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockDoubtGame) CpuPlay() {
	_m.Called()
}

// ResolveDoubt モック
func (_m *MockDoubtGame) ResolveDoubt(doubterIndices []int) {
	_m.Called(doubterIndices)
}

// SkipDoubt モック
func (_m *MockDoubtGame) SkipDoubt() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockDoubtGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockDoubtGame) GetPhase() domain.DoubtPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.DoubtPhase)
}

// IsHumanTurn モック
func (_m *MockDoubtGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetCurrentTurn モック
func (_m *MockDoubtGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayerCnt モック
func (_m *MockDoubtGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockDoubtGame) GetPlayer(i int) *domain.DoubtPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.DoubtPlayer); ok {
		return val
	}
	return nil
}

// GetTableCardCount モック
func (_m *MockDoubtGame) GetTableCardCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTableCards モック
func (_m *MockDoubtGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetLastAction モック
func (_m *MockDoubtGame) GetLastAction() *domain.DoubtAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.DoubtAction); ok {
		return val
	}
	return nil
}

// GetCpuDoubters モック
func (_m *MockDoubtGame) GetCpuDoubters() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockDoubtGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCpuActions モック
func (_m *MockDoubtGame) GetCpuActions() []*domain.DoubtCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.DoubtCpuAction); ok {
		return val
	}
	return nil
}

// GetHumanAction モック
func (_m *MockDoubtGame) GetHumanAction() *domain.DoubtCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.DoubtCpuAction); ok {
		return val
	}
	return nil
}

// GetLastDoubtResult モック
func (_m *MockDoubtGame) GetLastDoubtResult() *domain.DoubtDoubtResult {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.DoubtDoubtResult); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockDoubtGame) GetConfig() domain.DoubtConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DoubtConfig)
}

// SetConfig モック
func (_m *MockDoubtGame) SetConfig(cfg domain.DoubtConfig) {
	_m.Called(cfg)
}

// GetHumanProfile モック
func (_m *MockDoubtGame) GetHumanProfile() *domain.DoubtHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.DoubtHumanProfile); ok {
		return val
	}
	return nil
}

// ResetProfile モック
func (_m *MockDoubtGame) ResetProfile() {
	_m.Called()
}

// GetActionLog モック
func (_m *MockDoubtGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}

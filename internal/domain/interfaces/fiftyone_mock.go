//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFiftyOneGame フィフティワンゲームモック
type MockFiftyOneGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockFiftyOneGame) Reset() {
	_m.Called()
}

// SetConfig モック
func (_m *MockFiftyOneGame) SetConfig(cfg domain.FiftyOneConfig) {
	_m.Called(cfg)
}

// GetConfig モック
func (_m *MockFiftyOneGame) GetConfig() domain.FiftyOneConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.FiftyOneConfig)
}

// ExchangeOne モック
func (_m *MockFiftyOneGame) ExchangeOne(handIdx, tableIdx int) error {
	ret := _m.Called(handIdx, tableIdx)
	return ret.Error(0)
}

// ExchangeAll モック
func (_m *MockFiftyOneGame) ExchangeAll() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Stop モック
func (_m *MockFiftyOneGame) Stop() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockFiftyOneGame) CpuPlay() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetGameEndFlag モック
func (_m *MockFiftyOneGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockFiftyOneGame) GetPhase() domain.FiftyOnePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.FiftyOnePhase)
}

// IsHumanTurn モック
func (_m *MockFiftyOneGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerCnt モック
func (_m *MockFiftyOneGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockFiftyOneGame) GetPlayer(i int) *domain.FiftyOnePlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.FiftyOnePlayer); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockFiftyOneGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetWinnerIdx モック
func (_m *MockFiftyOneGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTableCards モック
func (_m *MockFiftyOneGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetStopCallerIdx モック
func (_m *MockFiftyOneGame) GetStopCallerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTurnNumber モック
func (_m *MockFiftyOneGame) GetTurnNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastAction モック
func (_m *MockFiftyOneGame) GetLastAction() string {
	ret := _m.Called()
	return ret.String(0)
}

// GetLastHandIdx モック
func (_m *MockFiftyOneGame) GetLastHandIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastTableIdx モック
func (_m *MockFiftyOneGame) GetLastTableIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockFiftyOneGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

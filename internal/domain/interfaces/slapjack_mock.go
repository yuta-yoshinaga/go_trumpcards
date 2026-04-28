//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSlapjackGame スラップジャックゲームモック
type MockSlapjackGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSlapjackGame) Reset() { _m.Called() }

// ResetWithConfig モック
func (_m *MockSlapjackGame) ResetWithConfig(cfg domain.SlapjackConfig) { _m.Called(cfg) }

// Step モック
func (_m *MockSlapjackGame) Step() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Slap モック
func (_m *MockSlapjackGame) Slap(playerIdx int) error {
	ret := _m.Called(playerIdx)
	return ret.Error(0)
}

// Tick モック
func (_m *MockSlapjackGame) Tick() domain.SlapjackPendingKind {
	ret := _m.Called()
	return ret.Get(0).(domain.SlapjackPendingKind)
}

// GetConfig モック
func (_m *MockSlapjackGame) GetConfig() domain.SlapjackConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SlapjackConfig)
}

// SetConfig モック
func (_m *MockSlapjackGame) SetConfig(cfg domain.SlapjackConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockSlapjackGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockSlapjackGame) GetPhase() domain.SlapjackPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SlapjackPhase)
}

// IsHumanTurn モック
func (_m *MockSlapjackGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerCnt モック
func (_m *MockSlapjackGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockSlapjackGame) GetPlayer(i int) *domain.SlapjackPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SlapjackPlayer)
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockSlapjackGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCenterPileSize モック
func (_m *MockSlapjackGame) GetCenterPileSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTopCard モック
func (_m *MockSlapjackGame) GetTopCard() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetCurrentTurnIdx モック
func (_m *MockSlapjackGame) GetCurrentTurnIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsTopJack モック
func (_m *MockSlapjackGame) IsTopJack() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPending モック
func (_m *MockSlapjackGame) GetPending() domain.SlapjackPending {
	ret := _m.Called()
	return ret.Get(0).(domain.SlapjackPending)
}

// GetLastEvent モック
func (_m *MockSlapjackGame) GetLastEvent() domain.SlapjackLastEvent {
	ret := _m.Called()
	return ret.Get(0).(domain.SlapjackLastEvent)
}

// GetActionLog モック
func (_m *MockSlapjackGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEgyptianRatscrewGame エジプシャン・ラットスクリューゲームモック
type MockEgyptianRatscrewGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockEgyptianRatscrewGame) Reset() { _m.Called() }

// ResetWithConfig モック
func (_m *MockEgyptianRatscrewGame) ResetWithConfig(cfg domain.EgyptianRatscrewConfig) {
	_m.Called(cfg)
}

// Step モック
func (_m *MockEgyptianRatscrewGame) Step() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Slap モック
func (_m *MockEgyptianRatscrewGame) Slap(playerIdx int) error {
	ret := _m.Called(playerIdx)
	return ret.Error(0)
}

// Tick モック
func (_m *MockEgyptianRatscrewGame) Tick() domain.EgyptianRatscrewPendingKind {
	ret := _m.Called()
	return ret.Get(0).(domain.EgyptianRatscrewPendingKind)
}

// GetConfig モック
func (_m *MockEgyptianRatscrewGame) GetConfig() domain.EgyptianRatscrewConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.EgyptianRatscrewConfig)
}

// SetConfig モック
func (_m *MockEgyptianRatscrewGame) SetConfig(cfg domain.EgyptianRatscrewConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockEgyptianRatscrewGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockEgyptianRatscrewGame) GetPhase() domain.EgyptianRatscrewPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.EgyptianRatscrewPhase)
}

// IsHumanTurn モック
func (_m *MockEgyptianRatscrewGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerCnt モック
func (_m *MockEgyptianRatscrewGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockEgyptianRatscrewGame) GetPlayer(i int) *domain.EgyptianRatscrewPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.EgyptianRatscrewPlayer)
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockEgyptianRatscrewGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCenterPileSize モック
func (_m *MockEgyptianRatscrewGame) GetCenterPileSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTopCard モック
func (_m *MockEgyptianRatscrewGame) GetTopCard() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetCurrentTurnIdx モック
func (_m *MockEgyptianRatscrewGame) GetCurrentTurnIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsTopFaceCard モック
func (_m *MockEgyptianRatscrewGame) IsTopFaceCard() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsSlappable モック
func (_m *MockEgyptianRatscrewGame) IsSlappable() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetChanceRemaining モック
func (_m *MockEgyptianRatscrewGame) GetChanceRemaining() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetChanceFromIdx モック
func (_m *MockEgyptianRatscrewGame) GetChanceFromIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPending モック
func (_m *MockEgyptianRatscrewGame) GetPending() domain.EgyptianRatscrewPending {
	ret := _m.Called()
	return ret.Get(0).(domain.EgyptianRatscrewPending)
}

// GetLastEvent モック
func (_m *MockEgyptianRatscrewGame) GetLastEvent() domain.EgyptianRatscrewLastEvent {
	ret := _m.Called()
	return ret.Get(0).(domain.EgyptianRatscrewLastEvent)
}

// GetActionLog モック
func (_m *MockEgyptianRatscrewGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

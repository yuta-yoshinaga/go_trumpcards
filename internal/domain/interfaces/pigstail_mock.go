//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPigsTailGame ぶたのしっぽゲームモック
type MockPigsTailGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPigsTailGame) Reset() {
	_m.Called()
}

// SetConfig モック
func (_m *MockPigsTailGame) SetConfig(config domain.PigsTailConfig) {
	_m.Called(config)
}

// GetGameEndFlag モック
func (_m *MockPigsTailGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockPigsTailGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerAction モック
func (_m *MockPigsTailGame) PlayerAction(actionIdx int) error {
	ret := _m.Called(actionIdx)
	return ret.Error(0)
}

// CpuAction モック
func (_m *MockPigsTailGame) CpuAction() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetPlayerCnt モック
func (_m *MockPigsTailGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockPigsTailGame) GetPlayer(i int) *domain.PigsTailPlayer {
	ret := _m.Called(i)
	return ret.Get(0).(*domain.PigsTailPlayer)
}

// GetCurrentTurn モック
func (_m *MockPigsTailGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLoserIdx モック
func (_m *MockPigsTailGame) GetLoserIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastDrawCard モック
func (_m *MockPigsTailGame) GetLastDrawCard() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetLastPenalty モック
func (_m *MockPigsTailGame) GetLastPenalty() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetCpuActions モック
func (_m *MockPigsTailGame) GetCpuActions() []*domain.PigsTailCpuAction {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.PigsTailCpuAction)
	}
	return nil
}

// GetHumanAction モック
func (_m *MockPigsTailGame) GetHumanAction() *domain.PigsTailCpuAction {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.PigsTailCpuAction)
	}
	return nil
}

// GetCircleCount モック
func (_m *MockPigsTailGame) GetCircleCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCenter モック
func (_m *MockPigsTailGame) GetCenter() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetCenterTopCard モック
func (_m *MockPigsTailGame) GetCenterTopCard() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetConfig モック
func (_m *MockPigsTailGame) GetConfig() domain.PigsTailConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PigsTailConfig)
}

// GetActionLog モック
func (_m *MockPigsTailGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

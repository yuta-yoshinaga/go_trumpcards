//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCuarentaGame クアレンタゲームのモック (testify/mock)。
type MockCuarentaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockCuarentaGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockCuarentaGame) NextRound() { _m.Called() }

// GetGameEndFlag モック
func (_m *MockCuarentaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockCuarentaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockCuarentaGame) PlayerPlay(handIdx int) error {
	ret := _m.Called(handIdx)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockCuarentaGame) CpuPlay() { _m.Called() }

// SetConfig モック
func (_m *MockCuarentaGame) SetConfig(config domain.CuarentaConfig) { _m.Called(config) }

// GetPlayerCnt モック
func (_m *MockCuarentaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockCuarentaGame) GetPlayer(i int) *domain.CuarentaPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.CuarentaPlayer); ok {
		return v
	}
	return nil
}

// GetTeamScore モック
func (_m *MockCuarentaGame) GetTeamScore(team int) int {
	ret := _m.Called(team)
	return ret.Int(0)
}

// GetTableCards モック
func (_m *MockCuarentaGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetLastCaptureIdx モック
func (_m *MockCuarentaGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanAction モック
func (_m *MockCuarentaGame) GetHumanAction() *domain.CuarentaAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.CuarentaAction); ok {
		return v
	}
	return nil
}

// GetCpuActions モック
func (_m *MockCuarentaGame) GetCpuActions() []*domain.CuarentaAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.CuarentaAction); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockCuarentaGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockCuarentaGame) GetConfig() domain.CuarentaConfig {
	ret := _m.Called()
	if v, ok := ret.Get(0).(domain.CuarentaConfig); ok {
		return v
	}
	return domain.CuarentaConfig{}
}

// GetPhase モック
func (_m *MockCuarentaGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastRoundDetail モック
func (_m *MockCuarentaGame) GetLastRoundDetail() *domain.CuarentaRoundDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.CuarentaRoundDetail); ok {
		return v
	}
	return nil
}

// GetRoundWinners モック
func (_m *MockCuarentaGame) GetRoundWinners() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetRemainingDeck モック
func (_m *MockCuarentaGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockCuarentaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

// GetTeamCapturedCount モック
func (m *MockCuarentaGame) GetTeamCapturedCount(team int) int { return m.Called(team).Int(0) }

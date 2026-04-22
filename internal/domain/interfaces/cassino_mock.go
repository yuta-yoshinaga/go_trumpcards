//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCassinoGame カシノゲームのモック (testify/mock)。
type MockCassinoGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockCassinoGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockCassinoGame) NextRound() { _m.Called() }

// GetGameEndFlag モック
func (_m *MockCassinoGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockCassinoGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerTake モック
func (_m *MockCassinoGame) PlayerTake(handIdx int, tableIdxs []int, buildIdxs []int) error {
	ret := _m.Called(handIdx, tableIdxs, buildIdxs)
	return ret.Error(0)
}

// PlayerBuild モック
func (_m *MockCassinoGame) PlayerBuild(handIdx int, tableIdxs []int, declaredValue int) error {
	ret := _m.Called(handIdx, tableIdxs, declaredValue)
	return ret.Error(0)
}

// PlayerTrail モック
func (_m *MockCassinoGame) PlayerTrail(handIdx int) error {
	ret := _m.Called(handIdx)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockCassinoGame) CpuPlay() { _m.Called() }

// SetConfig モック
func (_m *MockCassinoGame) SetConfig(config domain.CassinoConfig) { _m.Called(config) }

// GetPlayerCnt モック
func (_m *MockCassinoGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockCassinoGame) GetPlayer(i int) *domain.CassinoPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.CassinoPlayer); ok {
		return v
	}
	return nil
}

// GetTableCards モック
func (_m *MockCassinoGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetBuilds モック
func (_m *MockCassinoGame) GetBuilds() []*domain.CassinoBuild {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.CassinoBuild); ok {
		return v
	}
	return nil
}

// GetLastCaptureIdx モック
func (_m *MockCassinoGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanAction モック
func (_m *MockCassinoGame) GetHumanAction() *domain.CassinoAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.CassinoAction); ok {
		return v
	}
	return nil
}

// GetCpuActions モック
func (_m *MockCassinoGame) GetCpuActions() []*domain.CassinoAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.CassinoAction); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockCassinoGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockCassinoGame) GetConfig() domain.CassinoConfig {
	ret := _m.Called()
	if v, ok := ret.Get(0).(domain.CassinoConfig); ok {
		return v
	}
	return domain.CassinoConfig{}
}

// GetPhase モック
func (_m *MockCassinoGame) GetPhase() string {
	ret := _m.Called()
	return ret.String(0)
}

// GetLastRoundDetail モック
func (_m *MockCassinoGame) GetLastRoundDetail() *domain.CassinoScoreDetail {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.CassinoScoreDetail); ok {
		return v
	}
	return nil
}

// GetRoundWinners モック
func (_m *MockCassinoGame) GetRoundWinners() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetRemainingDeck モック
func (_m *MockCassinoGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPacksDealt モック
func (_m *MockCassinoGame) GetPacksDealt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockCassinoGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

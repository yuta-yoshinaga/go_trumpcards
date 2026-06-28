//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBeggarMyNeighbourGame Beggar-My-Neighbour ゲームモック
type MockBeggarMyNeighbourGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockBeggarMyNeighbourGame) Reset() { _m.Called() }

// Step モック
func (_m *MockBeggarMyNeighbourGame) Step() error {
	ret := _m.Called()
	return ret.Error(0)
}

// AutoPlay モック
func (_m *MockBeggarMyNeighbourGame) AutoPlay() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockBeggarMyNeighbourGame) GetConfig() domain.BeggarMyNeighbourConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BeggarMyNeighbourConfig)
}

// SetConfig モック
func (_m *MockBeggarMyNeighbourGame) SetConfig(cfg domain.BeggarMyNeighbourConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockBeggarMyNeighbourGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockBeggarMyNeighbourGame) GetPhase() domain.BeggarMyNeighbourPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BeggarMyNeighbourPhase)
}

// IsHumanTurn モック
func (_m *MockBeggarMyNeighbourGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerCnt モック
func (_m *MockBeggarMyNeighbourGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockBeggarMyNeighbourGame) GetPlayer(i int) *domain.BeggarMyNeighbourPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.BeggarMyNeighbourPlayer)
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockBeggarMyNeighbourGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockBeggarMyNeighbourGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPenaltyOwnerIdx モック
func (_m *MockBeggarMyNeighbourGame) GetPenaltyOwnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPenaltyRemaining モック
func (_m *MockBeggarMyNeighbourGame) GetPenaltyRemaining() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCentralPileSize モック
func (_m *MockBeggarMyNeighbourGame) GetCentralPileSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastCardPlayed モック
func (_m *MockBeggarMyNeighbourGame) GetLastCardPlayed() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetRoundsPlayed モック
func (_m *MockBeggarMyNeighbourGame) GetRoundsPlayed() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockBeggarMyNeighbourGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

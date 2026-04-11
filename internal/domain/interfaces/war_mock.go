//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWarGame 戦争ゲームモック
type MockWarGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockWarGame) Reset() { _m.Called() }

// Step モック
func (_m *MockWarGame) Step() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockWarGame) GetConfig() domain.WarConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.WarConfig)
}

// SetConfig モック
func (_m *MockWarGame) SetConfig(cfg domain.WarConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockWarGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockWarGame) GetPhase() domain.WarPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.WarPhase)
}

// IsHumanTurn モック
func (_m *MockWarGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPlayerCnt モック
func (_m *MockWarGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockWarGame) GetPlayer(i int) *domain.WarPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.WarPlayer)
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockWarGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayerRevealed モック
func (_m *MockWarGame) GetPlayerRevealed() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetCpuRevealed モック
func (_m *MockWarGame) GetCpuRevealed() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetWarPotSize モック
func (_m *MockWarGame) GetWarPotSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastWinnerIdx モック
func (_m *MockWarGame) GetLastWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastBurialCount モック
func (_m *MockWarGame) GetLastBurialCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRoundsPlayed モック
func (_m *MockWarGame) GetRoundsPlayed() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockWarGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

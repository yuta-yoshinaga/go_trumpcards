//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTablanetGame はタブラネット (Tablanet) のゲームモック。
type MockTablanetGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTablanetGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockTablanetGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockTablanetGame) PlayerPlay(handIdx int, tableIdxs []int) error {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockTablanetGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockTablanetGame) GetConfig() domain.TablanetConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TablanetConfig)
}

// SetConfig モック
func (_m *MockTablanetGame) SetConfig(cfg domain.TablanetConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockTablanetGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockTablanetGame) GetPhase() domain.TablanetPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TablanetPhase)
}

// IsHumanTurn モック
func (_m *MockTablanetGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCurrentTurn モック
func (_m *MockTablanetGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTableCards モック
func (_m *MockTablanetGame) GetTableCards() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetLastCaptureIdx モック
func (_m *MockTablanetGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRemainingDeck モック
func (_m *MockTablanetGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockTablanetGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinners モック
func (_m *MockTablanetGame) GetWinners() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetLastDealDetail モック
func (_m *MockTablanetGame) GetLastDealDetail() *domain.TablanetScoreDetail {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TablanetScoreDetail)
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockTablanetGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockTablanetGame) GetPlayer(i int) *domain.TablanetPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.TablanetPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockTablanetGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetCaptureOptions モック
func (_m *MockTablanetGame) GetCaptureOptions(playerIdx int) map[int][]int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.(map[int][]int)
	}
	return nil
}

// GetHint モック
func (_m *MockTablanetGame) GetHint() *domain.TablanetHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TablanetHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockTablanetGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

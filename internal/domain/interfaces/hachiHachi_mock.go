//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHachiHachiGame は八八 (Hachi-Hachi) のゲームモック。
type MockHachiHachiGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockHachiHachiGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockHachiHachiGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockHachiHachiGame) PlayerPlay(handIdx, fieldIdx int) error {
	ret := _m.Called(handIdx, fieldIdx)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockHachiHachiGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockHachiHachiGame) GetConfig() domain.HachiHachiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.HachiHachiConfig)
}

// SetConfig モック
func (_m *MockHachiHachiGame) SetConfig(cfg domain.HachiHachiConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockHachiHachiGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockHachiHachiGame) GetPhase() domain.HachiHachiPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.HachiHachiPhase)
}

// IsHumanTurn モック
func (_m *MockHachiHachiGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCurrentTurn モック
func (_m *MockHachiHachiGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetFieldCards モック
func (_m *MockHachiHachiGame) GetFieldCards() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetRemainingDeck モック
func (_m *MockHachiHachiGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockHachiHachiGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastRoundResult モック
func (_m *MockHachiHachiGame) GetLastRoundResult() *domain.HachiHachiRoundResult {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.HachiHachiRoundResult)
	}
	return nil
}

// GetWinner モック
func (_m *MockHachiHachiGame) GetWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetResult モック
func (_m *MockHachiHachiGame) GetResult() domain.HachiHachiResult {
	ret := _m.Called()
	return ret.Get(0).(domain.HachiHachiResult)
}

// GetPlayerCnt モック
func (_m *MockHachiHachiGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockHachiHachiGame) GetPlayer(i int) *domain.HachiHachiPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.HachiHachiPlayer)
	}
	return nil
}

// GetYaku モック
func (_m *MockHachiHachiGame) GetYaku(playerIdx int) ([]domain.HachiHachiYaku, int) {
	ret := _m.Called(playerIdx)
	var yakus []domain.HachiHachiYaku
	if v := ret.Get(0); v != nil {
		yakus = v.([]domain.HachiHachiYaku)
	}
	return yakus, ret.Int(1)
}

// GetPlayableIndices モック
func (_m *MockHachiHachiGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetCaptureOptions モック
func (_m *MockHachiHachiGame) GetCaptureOptions(playerIdx int) map[int][]int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.(map[int][]int)
	}
	return nil
}

// GetHint モック
func (_m *MockHachiHachiGame) GetHint() *domain.HachiHachiHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.HachiHachiHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockHachiHachiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

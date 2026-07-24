//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGutsGame はガッツ (Guts) のゲームモック。
type MockGutsGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockGutsGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockGutsGame) NextRound() { _m.Called() }

// Declare モック
func (_m *MockGutsGame) Declare(stay bool) error {
	ret := _m.Called(stay)
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockGutsGame) GetConfig() domain.GutsConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GutsConfig)
}

// SetConfig モック
func (_m *MockGutsGame) SetConfig(cfg domain.GutsConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockGutsGame) GetPhase() domain.GutsPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GutsPhase)
}

// GetGameEndFlag モック
func (_m *MockGutsGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockGutsGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockGutsGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCarryPot モック
func (_m *MockGutsGame) GetCarryPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCarryCount モック
func (_m *MockGutsGame) GetCarryCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAnte モック
func (_m *MockGutsGame) GetAnte() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerIdx モック
func (_m *MockGutsGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMatchWinnerIdx モック
func (_m *MockGutsGame) GetMatchWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetResult モック
func (_m *MockGutsGame) GetResult() domain.GutsResult {
	ret := _m.Called()
	return ret.Get(0).(domain.GutsResult)
}

// GetMatchers モック
func (_m *MockGutsGame) GetMatchers() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// IsMatcher モック
func (_m *MockGutsGame) IsMatcher(idx int) bool {
	ret := _m.Called(idx)
	return ret.Get(0).(bool)
}

// GetPlayerCnt モック
func (_m *MockGutsGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockGutsGame) GetPlayer(i int) *domain.GutsPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.GutsPlayer)
	}
	return nil
}

// GetChips モック
func (_m *MockGutsGame) GetChips() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockGutsGame) GetHint() *domain.GutsHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.GutsHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockGutsGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

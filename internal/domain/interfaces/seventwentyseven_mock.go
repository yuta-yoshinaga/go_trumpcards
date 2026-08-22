//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSevenTwentySevenGame はセブン・トゥエンティセブン (SevenTwentySeven) のゲームモック。
type MockSevenTwentySevenGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSevenTwentySevenGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockSevenTwentySevenGame) NextRound() { _m.Called() }

// Declare モック
func (_m *MockSevenTwentySevenGame) Declare(stay bool) error {
	ret := _m.Called(stay)
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockSevenTwentySevenGame) GetConfig() domain.SevenTwentySevenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SevenTwentySevenConfig)
}

// SetConfig モック
func (_m *MockSevenTwentySevenGame) SetConfig(cfg domain.SevenTwentySevenConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockSevenTwentySevenGame) GetPhase() domain.SevenTwentySevenPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SevenTwentySevenPhase)
}

// GetGameEndFlag モック
func (_m *MockSevenTwentySevenGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockSevenTwentySevenGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockSevenTwentySevenGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCarryPot モック
func (_m *MockSevenTwentySevenGame) GetCarryPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCarryCount モック
func (_m *MockSevenTwentySevenGame) GetCarryCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAnte モック
func (_m *MockSevenTwentySevenGame) GetAnte() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerIdx モック
func (_m *MockSevenTwentySevenGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMatchWinnerIdx モック
func (_m *MockSevenTwentySevenGame) GetMatchWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetResult モック
func (_m *MockSevenTwentySevenGame) GetResult() domain.SevenTwentySevenResult {
	ret := _m.Called()
	return ret.Get(0).(domain.SevenTwentySevenResult)
}

// GetMatchers モック
func (_m *MockSevenTwentySevenGame) GetMatchers() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// IsMatcher モック
func (_m *MockSevenTwentySevenGame) IsMatcher(idx int) bool {
	ret := _m.Called(idx)
	return ret.Get(0).(bool)
}

// GetPlayerCnt モック
func (_m *MockSevenTwentySevenGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSevenTwentySevenGame) GetPlayer(i int) *domain.SevenTwentySevenPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SevenTwentySevenPlayer)
	}
	return nil
}

// GetChips モック
func (_m *MockSevenTwentySevenGame) GetChips() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockSevenTwentySevenGame) GetHint() *domain.SevenTwentySevenHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.SevenTwentySevenHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockSevenTwentySevenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

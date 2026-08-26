//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCometGame はコメットのゲームモック。
type MockCometGame struct {
	mock.Mock
}

// GetActionLog モック
func (_m *MockCometGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v, _ := ret.Get(0).([]*domain.ActionLogEntry)
	return v
}

// Reset モック
func (_m *MockCometGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockCometGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockCometGame) PlayerPlay(handIdx int) error {
	ret := _m.Called(handIdx)
	err, _ := ret.Get(0).(error)
	return err
}

// PlayerPass モック
func (_m *MockCometGame) PlayerPass() error {
	ret := _m.Called()
	err, _ := ret.Get(0).(error)
	return err
}

// CpuPlay モック
func (_m *MockCometGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockCometGame) GetConfig() domain.CometConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CometConfig)
}

// SetConfig モック
func (_m *MockCometGame) SetConfig(cfg domain.CometConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockCometGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockCometGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockCometGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPile モック
func (_m *MockCometGame) GetPile() []*domain.Card {
	ret := _m.Called()
	v, _ := ret.Get(0).([]*domain.Card)
	return v
}

// GetNeed モック
func (_m *MockCometGame) GetNeed() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeadCount モック
func (_m *MockCometGame) GetDeadCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockCometGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockCometGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockCometGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastPlayerIdx モック
func (_m *MockCometGame) GetLastPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// PlayableIdxs モック
func (_m *MockCometGame) PlayableIdxs(seat int) []int {
	ret := _m.Called(seat)
	v, _ := ret.Get(0).([]int)
	return v
}

// GetPlayerCnt モック
func (_m *MockCometGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockCometGame) GetPlayer(i int) *domain.CometPlayer {
	ret := _m.Called(i)
	v, _ := ret.Get(0).(*domain.CometPlayer)
	return v
}

// GetLastResult モック
func (_m *MockCometGame) GetLastResult() *domain.CometRoundResult {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.CometRoundResult)
	return v
}

// GetWinnerIdx モック
func (_m *MockCometGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockCometGame) GetHint() *domain.CometHint {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.CometHint)
	return v
}

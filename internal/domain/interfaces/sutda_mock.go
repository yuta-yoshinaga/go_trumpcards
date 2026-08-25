//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSutdaGame はソッタのゲームモック。
type MockSutdaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSutdaGame) Reset() {
	_m.Called()
}

// NextHand モック
func (_m *MockSutdaGame) NextHand() {
	_m.Called()
}

// PlayerAction モック
func (_m *MockSutdaGame) PlayerAction(action string) error {
	ret := _m.Called(action)
	if err, ok := ret.Get(0).(error); ok {
		return err
	}
	return nil
}

// CpuAction モック
func (_m *MockSutdaGame) CpuAction() {
	_m.Called()
}

// GetConfig モック
func (_m *MockSutdaGame) GetConfig() domain.SutdaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SutdaConfig)
}

// SetConfig モック
func (_m *MockSutdaGame) SetConfig(cfg domain.SutdaConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockSutdaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockSutdaGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockSutdaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetHandNumber モック
func (_m *MockSutdaGame) GetHandNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockSutdaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockSutdaGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockSutdaGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBet モック
func (_m *MockSutdaGame) GetCurrentBet() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRaiseCount モック
func (_m *MockSutdaGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCallAmount モック
func (_m *MockSutdaGame) GetCallAmount(playerIdx int) int {
	ret := _m.Called(playerIdx)
	return ret.Get(0).(int)
}

// CanRaise モック
func (_m *MockSutdaGame) CanRaise(playerIdx int) bool {
	ret := _m.Called(playerIdx)
	return ret.Get(0).(bool)
}

// GetHandOf モック
func (_m *MockSutdaGame) GetHandOf(playerIdx int) domain.SutdaHand {
	ret := _m.Called(playerIdx)
	return ret.Get(0).(domain.SutdaHand)
}

// GetPlayerCnt モック
func (_m *MockSutdaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSutdaGame) GetPlayer(i int) *domain.SutdaPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.SutdaPlayer); ok {
		return v
	}
	return nil
}

// GetLastResult モック
func (_m *MockSutdaGame) GetLastResult() *domain.SutdaHandResult {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.SutdaHandResult); ok {
		return v
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockSutdaGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockSutdaGame) GetHint() *domain.SutdaHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.SutdaHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockSutdaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

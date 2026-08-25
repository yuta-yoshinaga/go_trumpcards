//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCirullaGame はチルッラのゲームモック。
type MockCirullaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockCirullaGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockCirullaGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockCirullaGame) PlayerPlay(handIdx int, captureIdxs []int) error {
	ret := _m.Called(handIdx, captureIdxs)
	if err, ok := ret.Get(0).(error); ok {
		return err
	}
	return nil
}

// CpuPlay モック
func (_m *MockCirullaGame) CpuPlay() {
	_m.Called()
}

// GetConfig モック
func (_m *MockCirullaGame) GetConfig() domain.CirullaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CirullaConfig)
}

// SetConfig モック
func (_m *MockCirullaGame) SetConfig(cfg domain.CirullaConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockCirullaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockCirullaGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockCirullaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetTable モック
func (_m *MockCirullaGame) GetTable() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetRoundNumber モック
func (_m *MockCirullaGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockCirullaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockCirullaGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastCapturer モック
func (_m *MockCirullaGame) GetLastCapturer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastBonus モック
func (_m *MockCirullaGame) GetLastBonus() []string {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]string); ok {
		return v
	}
	return nil
}

// GetDeckRemaining モック
func (_m *MockCirullaGame) GetDeckRemaining() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCaptureOptions モック
func (_m *MockCirullaGame) GetCaptureOptions(playerIdx, handIdx int) [][]int {
	ret := _m.Called(playerIdx, handIdx)
	if v, ok := ret.Get(0).([][]int); ok {
		return v
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockCirullaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockCirullaGame) GetPlayer(i int) *domain.CirullaPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.CirullaPlayer); ok {
		return v
	}
	return nil
}

// GetLastResult モック
func (_m *MockCirullaGame) GetLastResult() *domain.CirullaRoundResult {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.CirullaRoundResult); ok {
		return v
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockCirullaGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockCirullaGame) GetHint() *domain.CirullaHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.CirullaHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockCirullaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

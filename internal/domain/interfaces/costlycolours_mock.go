//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCostlyColoursGame はコストリー・カラーズのゲームモック。
type MockCostlyColoursGame struct {
	mock.Mock
}

// GetActionLog モック
func (_m *MockCostlyColoursGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v, _ := ret.Get(0).([]*domain.ActionLogEntry)
	return v
}

// Reset モック
func (_m *MockCostlyColoursGame) Reset() { _m.Called() }

// NextDeal モック
func (_m *MockCostlyColoursGame) NextDeal() { _m.Called() }

// PlayerMog モック
func (_m *MockCostlyColoursGame) PlayerMog(accept bool) error {
	ret := _m.Called(accept)
	err, _ := ret.Get(0).(error)
	return err
}

// PlayerPlay モック
func (_m *MockCostlyColoursGame) PlayerPlay(handIdx int) error {
	ret := _m.Called(handIdx)
	err, _ := ret.Get(0).(error)
	return err
}

// CpuAct モック
func (_m *MockCostlyColoursGame) CpuAct() { _m.Called() }

// GetConfig モック
func (_m *MockCostlyColoursGame) GetConfig() domain.CostlyColoursConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CostlyColoursConfig)
}

// SetConfig モック
func (_m *MockCostlyColoursGame) SetConfig(cfg domain.CostlyColoursConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockCostlyColoursGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockCostlyColoursGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockCostlyColoursGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetTurnUp モック
func (_m *MockCostlyColoursGame) GetTurnUp() *domain.Card {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.Card)
	return v
}

// GetPile モック
func (_m *MockCostlyColoursGame) GetPile() []*domain.Card {
	ret := _m.Called()
	v, _ := ret.Get(0).([]*domain.Card)
	return v
}

// GetTotal モック
func (_m *MockCostlyColoursGame) GetTotal() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWentOut モック
func (_m *MockCostlyColoursGame) GetWentOut() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealNumber モック
func (_m *MockCostlyColoursGame) GetDealNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockCostlyColoursGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockCostlyColoursGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// PlayableIdxs モック
func (_m *MockCostlyColoursGame) PlayableIdxs(seat int) []int {
	ret := _m.Called(seat)
	v, _ := ret.Get(0).([]int)
	return v
}

// GetPlayerCnt モック
func (_m *MockCostlyColoursGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockCostlyColoursGame) GetPlayer(i int) *domain.CostlyColoursPlayer {
	ret := _m.Called(i)
	v, _ := ret.Get(0).(*domain.CostlyColoursPlayer)
	return v
}

// GetLastResult モック
func (_m *MockCostlyColoursGame) GetLastResult() *domain.CostlyColoursDealResult {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.CostlyColoursDealResult)
	return v
}

// GetWinnerIdx モック
func (_m *MockCostlyColoursGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockCostlyColoursGame) GetHint() *domain.CostlyColoursHint {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.CostlyColoursHint)
	return v
}

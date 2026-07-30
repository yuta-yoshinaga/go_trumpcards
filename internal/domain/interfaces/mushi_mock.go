//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMushiGame 虫 ゲームモック
type MockMushiGame struct {
	mock.Mock
}

func (_m *MockMushiGame) Reset() { _m.Called() }

func (_m *MockMushiGame) PlayCard(player, handIdx int) error {
	ret := _m.Called(player, handIdx)
	return ret.Error(0)
}

func (_m *MockMushiGame) SelectCapture(player, fieldIdx int) error {
	ret := _m.Called(player, fieldIdx)
	return ret.Error(0)
}

func (_m *MockMushiGame) NextRound() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockMushiGame) MushiCpuDecide(idx int) domain.MushiCpuAction {
	ret := _m.Called(idx)
	return ret.Get(0).(domain.MushiCpuAction)
}

func (_m *MockMushiGame) GetConfig() domain.MushiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MushiConfig)
}

func (_m *MockMushiGame) SetConfig(cfg domain.MushiConfig) { _m.Called(cfg) }

func (_m *MockMushiGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMushiGame) GetPhase() domain.MushiPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MushiPhase)
}

func (_m *MockMushiGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMushiGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMushiGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMushiGame) GetField() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockMushiGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMushiGame) GetCaptured(idx int) []*domain.Card {
	ret := _m.Called(idx)
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockMushiGame) GetPlayers() []*domain.MushiPlayer {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.MushiPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockMushiGame) GetPlayer(idx int) *domain.MushiPlayer {
	ret := _m.Called(idx)
	if v, ok := ret.Get(0).(*domain.MushiPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockMushiGame) GetScore(idx int) int {
	ret := _m.Called(idx)
	return ret.Int(0)
}

func (_m *MockMushiGame) GetRoundResult(idx int) int {
	ret := _m.Called(idx)
	return ret.Int(0)
}

func (_m *MockMushiGame) GetPendingCard() *domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockMushiGame) GetSelectableIndices() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockMushiGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMushiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

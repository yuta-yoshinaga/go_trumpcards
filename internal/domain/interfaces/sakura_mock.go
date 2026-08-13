//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSakuraGame はさくら (肥後花) のゲームモック。
type MockSakuraGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSakuraGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockSakuraGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockSakuraGame) PlayerPlay(handIdx, fieldIdx int) error {
	ret := _m.Called(handIdx, fieldIdx)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSakuraGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockSakuraGame) GetConfig() domain.SakuraConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SakuraConfig)
}

// SetConfig モック
func (_m *MockSakuraGame) SetConfig(cfg domain.SakuraConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockSakuraGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPhase モック
func (_m *MockSakuraGame) GetPhase() domain.SakuraPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SakuraPhase)
}

// IsHumanTurn モック
func (_m *MockSakuraGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// HumanSeat モック
func (_m *MockSakuraGame) HumanSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTurn モック
func (_m *MockSakuraGame) GetTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDealer モック
func (_m *MockSakuraGame) GetDealer() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetField モック
func (_m *MockSakuraGame) GetField() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetStockCount モック
func (_m *MockSakuraGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRound モック
func (_m *MockSakuraGame) GetRound() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetWinner モック
func (_m *MockSakuraGame) GetWinner() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastResult モック
func (_m *MockSakuraGame) GetLastResult() *domain.SakuraRoundResult {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.SakuraRoundResult); ok {
		return v
	}
	return nil
}

// GetPlayerCnt モック
func (_m *MockSakuraGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockSakuraGame) GetPlayer(i int) *domain.SakuraPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.SakuraPlayer); ok {
		return v
	}
	return nil
}

// GetValidFieldIndices モック
func (_m *MockSakuraGame) GetValidFieldIndices() map[int][]int {
	ret := _m.Called()
	if v, ok := ret.Get(0).(map[int][]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockSakuraGame) GetHint() domain.SakuraHint {
	ret := _m.Called()
	return ret.Get(0).(domain.SakuraHint)
}

// GetActionLog モック
func (_m *MockSakuraGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}

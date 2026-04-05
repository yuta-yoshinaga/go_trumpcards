//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGolfGame ゴルフソリティアゲームモック
type MockGolfGame struct {
	mock.Mock
}

func (_m *MockGolfGame) Reset() {
	_m.Called()
}

func (_m *MockGolfGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockGolfGame) Remove(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockGolfGame) GiveUp() {
	_m.Called()
}

func (_m *MockGolfGame) GetHint() *domain.GolfHint {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.(*domain.GolfHint)
}

func (_m *MockGolfGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockGolfGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

func (_m *MockGolfGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockGolfGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockGolfGame) GetPhase() domain.GolfPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GolfPhase)
}

func (_m *MockGolfGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockGolfGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockGolfGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.([]*domain.Card)
}

func (_m *MockGolfGame) GetLayout() [domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard {
	ret := _m.Called()
	return ret.Get(0).([domain.GolfColCnt][domain.GolfRowCnt]*domain.GolfCard)
}

func (_m *MockGolfGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.([]*domain.ActionLogEntry)
}

func (_m *MockGolfGame) IsExposed(col, row int) bool {
	ret := _m.Called(col, row)
	return ret.Get(0).(bool)
}

func (_m *MockGolfGame) AllRemoved() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

func (_m *MockGolfGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

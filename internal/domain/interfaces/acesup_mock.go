//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAcesUpGame エースアップゲームモック
type MockAcesUpGame struct {
	mock.Mock
}

func (_m *MockAcesUpGame) Reset() {
	_m.Called()
}

func (_m *MockAcesUpGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAcesUpGame) Remove(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockAcesUpGame) Move(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockAcesUpGame) GiveUp() {
	_m.Called()
}

func (_m *MockAcesUpGame) GetHint() *domain.AcesUpHint {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.(*domain.AcesUpHint)
}

func (_m *MockAcesUpGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAcesUpGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

func (_m *MockAcesUpGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAcesUpGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockAcesUpGame) GetPhase() domain.AcesUpPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.AcesUpPhase)
}

func (_m *MockAcesUpGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockAcesUpGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockAcesUpGame) GetDiscardCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDiscardTop mocks the GetDiscardTop call.
func (_m *MockAcesUpGame) GetDiscardTop() *domain.Card {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.(*domain.Card)
}

func (_m *MockAcesUpGame) GetColumns() [domain.AcesUpColCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.AcesUpColCnt][]*domain.Card)
}

func (_m *MockAcesUpGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.([]*domain.ActionLogEntry)
}

func (_m *MockAcesUpGame) CanRemove(col int) bool {
	ret := _m.Called(col)
	return ret.Get(0).(bool)
}

func (_m *MockAcesUpGame) CanMove(col int) bool {
	ret := _m.Called(col)
	return ret.Get(0).(bool)
}

func (_m *MockAcesUpGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockAcesUpGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

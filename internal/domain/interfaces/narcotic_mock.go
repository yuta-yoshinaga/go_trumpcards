//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNarcoticGame ナルコティックゲームモック
type MockNarcoticGame struct {
	mock.Mock
}

func (_m *MockNarcoticGame) Reset() {
	_m.Called()
}

func (_m *MockNarcoticGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockNarcoticGame) Remove() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockNarcoticGame) Move(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockNarcoticGame) GiveUp() {
	_m.Called()
}

func (_m *MockNarcoticGame) GetHint() *domain.NarcoticHint {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.(*domain.NarcoticHint)
}

func (_m *MockNarcoticGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockNarcoticGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

func (_m *MockNarcoticGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockNarcoticGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockNarcoticGame) GetPhase() domain.NarcoticPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.NarcoticPhase)
}

func (_m *MockNarcoticGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockNarcoticGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockNarcoticGame) GetDiscardCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDiscardTop mocks the GetDiscardTop call.
func (_m *MockNarcoticGame) GetDiscardTop() *domain.Card {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.(*domain.Card)
}

func (_m *MockNarcoticGame) GetColumns() [domain.NarcoticColCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.NarcoticColCnt][]*domain.Card)
}

func (_m *MockNarcoticGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.([]*domain.ActionLogEntry)
}

func (_m *MockNarcoticGame) CanRemoveSet() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

func (_m *MockNarcoticGame) MoveTarget(col int) int {
	ret := _m.Called(col)
	return ret.Get(0).(int)
}

func (_m *MockNarcoticGame) Redeal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockNarcoticGame) GetRedealCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockNarcoticGame) CanMove(col int) bool {
	ret := _m.Called(col)
	return ret.Get(0).(bool)
}

func (_m *MockNarcoticGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockNarcoticGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

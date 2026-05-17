//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEightOffGame エイトオフゲームモック
type MockEightOffGame struct {
	mock.Mock
}

func (_m *MockEightOffGame) Reset() {
	_m.Called()
}

func (_m *MockEightOffGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockEightOffGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockEightOffGame) MoveTableauToFreeCell(col, cell int) error {
	ret := _m.Called(col, cell)
	return ret.Error(0)
}

func (_m *MockEightOffGame) MoveFreeCellToTableau(cell, col int) error {
	ret := _m.Called(cell, col)
	return ret.Error(0)
}

func (_m *MockEightOffGame) MoveFreeCellToFoundation(cell int) error {
	ret := _m.Called(cell)
	return ret.Error(0)
}

func (_m *MockEightOffGame) GiveUp() {
	_m.Called()
}

func (_m *MockEightOffGame) GetHint() *domain.EightOffHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.EightOffHint)
}

func (_m *MockEightOffGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockEightOffGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockEightOffGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockEightOffGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockEightOffGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockEightOffGame) GetPhase() domain.EightOffPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.EightOffPhase)
}

func (_m *MockEightOffGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockEightOffGame) GetTableau() [domain.EightOffTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.EightOffTableauCnt][]*domain.Card)
}

func (_m *MockEightOffGame) GetFreeCells() [domain.EightOffCellCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.EightOffCellCnt]*domain.Card)
}

func (_m *MockEightOffGame) GetFoundation() [domain.EightOffFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.EightOffFoundationCnt][]*domain.Card)
}

func (_m *MockEightOffGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockEightOffGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockEightOffGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

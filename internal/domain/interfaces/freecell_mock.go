//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFreeCellGame フリーセルゲームモック
type MockFreeCellGame struct {
	mock.Mock
}

func (_m *MockFreeCellGame) Reset() {
	_m.Called()
}

func (_m *MockFreeCellGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockFreeCellGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockFreeCellGame) MoveTableauToFreeCell(col, cell int) error {
	ret := _m.Called(col, cell)
	return ret.Error(0)
}

func (_m *MockFreeCellGame) MoveFreeCellToTableau(cell, col int) error {
	ret := _m.Called(cell, col)
	return ret.Error(0)
}

func (_m *MockFreeCellGame) MoveFreeCellToFoundation(cell int) error {
	ret := _m.Called(cell)
	return ret.Error(0)
}

func (_m *MockFreeCellGame) GiveUp() {
	_m.Called()
}

func (_m *MockFreeCellGame) GetHint() *domain.FreeCellHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.FreeCellHint)
}

func (_m *MockFreeCellGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFreeCellGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFreeCellGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFreeCellGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFreeCellGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockFreeCellGame) GetPhase() domain.FreeCellPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.FreeCellPhase)
}

func (_m *MockFreeCellGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFreeCellGame) GetTableau() [domain.FreeCellTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.FreeCellTableauCnt][]*domain.Card)
}

func (_m *MockFreeCellGame) GetFreeCells() [domain.FreeCellCellCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.FreeCellCellCnt]*domain.Card)
}

func (_m *MockFreeCellGame) GetMaxMovableCards() int {
	args := _m.Called()
	return args.Int(0)
}

func (_m *MockFreeCellGame) GetMaxMovableCardsToEmptyColumn() int {
	args := _m.Called()
	return args.Int(0)
}

func (_m *MockFreeCellGame) GetFoundation() [domain.FreeCellFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.FreeCellFoundationCnt][]*domain.Card)
}

func (_m *MockFreeCellGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockFreeCellGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockFreeCellGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

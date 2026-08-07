//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPenguinGame ペンギンゲームモック
type MockPenguinGame struct {
	mock.Mock
}

func (_m *MockPenguinGame) Reset() {
	_m.Called()
}

func (_m *MockPenguinGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockPenguinGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockPenguinGame) MoveTableauToFreeCell(col, cell int) error {
	ret := _m.Called(col, cell)
	return ret.Error(0)
}

func (_m *MockPenguinGame) MoveFreeCellToTableau(cell, col int) error {
	ret := _m.Called(cell, col)
	return ret.Error(0)
}

func (_m *MockPenguinGame) MoveFreeCellToFoundation(cell int) error {
	ret := _m.Called(cell)
	return ret.Error(0)
}

func (_m *MockPenguinGame) GiveUp() {
	_m.Called()
}

func (_m *MockPenguinGame) GetHint() *domain.PenguinHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.PenguinHint)
}

func (_m *MockPenguinGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPenguinGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPenguinGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPenguinGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPenguinGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockPenguinGame) GetPhase() domain.PenguinPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.PenguinPhase)
}

func (_m *MockPenguinGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPenguinGame) GetTableau() [domain.PenguinTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.PenguinTableauCnt][]*domain.Card)
}

func (_m *MockPenguinGame) GetFreeCells() [domain.PenguinCellCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.PenguinCellCnt]*domain.Card)
}

func (_m *MockPenguinGame) GetMaxMovableCards() int {
	args := _m.Called()
	return args.Int(0)
}

func (_m *MockPenguinGame) GetMaxMovableCardsToEmptyColumn() int {
	args := _m.Called()
	return args.Int(0)
}

func (_m *MockPenguinGame) GetFoundation() [domain.PenguinFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.PenguinFoundationCnt][]*domain.Card)
}

func (_m *MockPenguinGame) GetBaseRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPenguinGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockPenguinGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockPenguinGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

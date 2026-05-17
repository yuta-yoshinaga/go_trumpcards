//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockEightOffInteractor エイトオフインタラクターモック
type MockEightOffInteractor struct {
	mock.Mock
}

func (_m *MockEightOffInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) MoveTableauToFreeCell(col, cell int) string {
	ret := _m.Called(col, cell)
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) MoveFreeCellToTableau(cell, col int) string {
	ret := _m.Called(cell, col)
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) MoveFreeCellToFoundation(cell int) string {
	ret := _m.Called(cell)
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockEightOffInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockEightOffInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

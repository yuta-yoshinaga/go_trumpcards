//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockPenguinInteractor ペンギンインタラクターモック
type MockPenguinInteractor struct {
	mock.Mock
}

func (_m *MockPenguinInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) MoveTableauToFreeCell(col, cell int) string {
	ret := _m.Called(col, cell)
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) MoveFreeCellToTableau(cell, col int) string {
	ret := _m.Called(cell, col)
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) MoveFreeCellToFoundation(cell int) string {
	ret := _m.Called(cell)
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPenguinInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPenguinInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

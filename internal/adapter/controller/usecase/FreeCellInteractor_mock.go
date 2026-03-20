package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockFreeCellInteractor フリーセルインタラクターモック
type MockFreeCellInteractor struct {
	mock.Mock
}

func (_m *MockFreeCellInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) MoveTableauToFreeCell(col, cell int) string {
	ret := _m.Called(col, cell)
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) MoveFreeCellToTableau(cell, col int) string {
	ret := _m.Called(cell, col)
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) MoveFreeCellToFoundation(cell int) string {
	ret := _m.Called(cell)
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFreeCellInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockSeahavenTowersInteractor シーヘイブンタワーズインタラクターモック
type MockSeahavenTowersInteractor struct {
	mock.Mock
}

func (_m *MockSeahavenTowersInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) MoveTableauToFreeCell(col, cell int) string {
	ret := _m.Called(col, cell)
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) MoveFreeCellToTableau(cell, col int) string {
	ret := _m.Called(cell, col)
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) MoveFreeCellToFoundation(cell int) string {
	ret := _m.Called(cell)
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSeahavenTowersInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSeahavenTowersInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

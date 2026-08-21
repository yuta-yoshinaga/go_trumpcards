//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockStalactitesInteractor フリーセルインタラクターモック
type MockStalactitesInteractor struct {
	mock.Mock
}

func (_m *MockStalactitesInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) MoveTableauToStalactites(col, cell int) string {
	ret := _m.Called(col, cell)
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) MoveStalactitesToTableau(cell, col int) string {
	ret := _m.Called(cell, col)
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) MoveStalactitesToFoundation(cell int) string {
	ret := _m.Called(cell)
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStalactitesInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockStalactitesInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

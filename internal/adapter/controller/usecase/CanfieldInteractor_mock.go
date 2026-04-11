//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockCanfieldInteractor キャンフィールドインタラクターモック
type MockCanfieldInteractor struct {
	mock.Mock
}

func (_m *MockCanfieldInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) MoveWasteToTableau(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) MoveReserveToTableau(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) MoveReserveToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCanfieldInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCanfieldInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

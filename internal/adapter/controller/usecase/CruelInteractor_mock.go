//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockCruelInteractor クルーエルインタラクターモック
type MockCruelInteractor struct {
	mock.Mock
}

func (_m *MockCruelInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	ret := _m.Called(fromCol, toCol)
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) Shift() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

func (_m *MockCruelInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

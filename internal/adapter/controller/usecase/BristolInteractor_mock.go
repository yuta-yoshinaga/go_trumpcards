//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockBristolInteractor ブリストルインタラクターモック
type MockBristolInteractor struct {
	mock.Mock
}

func (_m *MockBristolInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	ret := _m.Called(fromCol, toCol)
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) MoveFanToTableau(fanIdx, toCol int) string {
	ret := _m.Called(fanIdx, toCol)
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) MoveFanToFoundation(fanIdx int) string {
	ret := _m.Called(fanIdx)
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBristolInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBristolInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

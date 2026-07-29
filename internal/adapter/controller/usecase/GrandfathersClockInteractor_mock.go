//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockGrandfathersClockInteractor グランドファーザーズ・クロック インタラクターモック
type MockGrandfathersClockInteractor struct {
	mock.Mock
}

func (_m *MockGrandfathersClockInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGrandfathersClockInteractor) MoveTableauToFoundation(col, fIdx int) string {
	ret := _m.Called(col, fIdx)
	return ret.Get(0).(string)
}

func (_m *MockGrandfathersClockInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	ret := _m.Called(fromCol, toCol)
	return ret.Get(0).(string)
}

func (_m *MockGrandfathersClockInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGrandfathersClockInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGrandfathersClockInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGrandfathersClockInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGrandfathersClockInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGrandfathersClockInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGrandfathersClockInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

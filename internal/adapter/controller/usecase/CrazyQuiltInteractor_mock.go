//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockCrazyQuiltInteractor クレイジーキルト インタラクターモック
type MockCrazyQuiltInteractor struct {
	mock.Mock
}

func (_m *MockCrazyQuiltInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) MoveQuiltToFoundation(idx int) string {
	ret := _m.Called(idx)
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) MoveQuiltToWaste(idx int) string {
	ret := _m.Called(idx)
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrazyQuiltInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCrazyQuiltInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

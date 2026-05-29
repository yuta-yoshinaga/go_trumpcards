//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockAcesUpInteractor エースアップインタラクターモック
type MockAcesUpInteractor struct {
	mock.Mock
}

func (_m *MockAcesUpInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAcesUpInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAcesUpInteractor) Remove(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockAcesUpInteractor) Move(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockAcesUpInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAcesUpInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAcesUpInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAcesUpInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAcesUpInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockAcesUpInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

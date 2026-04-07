//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockTriPeaksInteractor トリピークスインタラクターモック
type MockTriPeaksInteractor struct {
	mock.Mock
}

func (_m *MockTriPeaksInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksInteractor) Remove(row, col int) string {
	ret := _m.Called(row, col)
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTriPeaksInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTriPeaksInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

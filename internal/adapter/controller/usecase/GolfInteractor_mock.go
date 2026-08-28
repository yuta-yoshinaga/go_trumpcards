//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockGolfInteractor ゴルフソリティアインタラクターモック
type MockGolfInteractor struct {
	mock.Mock
}

func (_m *MockGolfInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGolfInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGolfInteractor) Remove(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockGolfInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGolfInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGolfInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGolfInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGolfInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGolfInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

func (_m *MockGolfInteractor) ResetNineHole() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

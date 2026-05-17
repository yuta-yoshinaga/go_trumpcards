//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockGapsInteractor はGapsInteractorIFのテスト用モック。
type MockGapsInteractor struct {
	mock.Mock
}

func (_m *MockGapsInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGapsInteractor) Move(fromRow, fromCol, toRow, toCol int) string {
	ret := _m.Called(fromRow, fromCol, toRow, toCol)
	return ret.Get(0).(string)
}

func (_m *MockGapsInteractor) Redeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGapsInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGapsInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

func (_m *MockGapsInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGapsInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGapsInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockGapsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

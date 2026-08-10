//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockCribbageSquaresInteractor はクリベッジ・スクエアズインタラクターのモック。
type MockCribbageSquaresInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCribbageSquaresInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Place モック
func (_m *MockCribbageSquaresInteractor) Place(row, col int) string {
	ret := _m.Called(row, col)
	return ret.Get(0).(string)
}

// Undo モック
func (_m *MockCribbageSquaresInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GiveUp モック
func (_m *MockCribbageSquaresInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockCribbageSquaresInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCribbageSquaresInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCribbageSquaresInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockPokerSquaresInteractor はポーカー・スクエアズインタラクターのモック。
type MockPokerSquaresInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPokerSquaresInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Place モック
func (_m *MockPokerSquaresInteractor) Place(row, col int) string {
	ret := _m.Called(row, col)
	return ret.Get(0).(string)
}

// Undo モック
func (_m *MockPokerSquaresInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GiveUp モック
func (_m *MockPokerSquaresInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockPokerSquaresInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPokerSquaresInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPokerSquaresInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

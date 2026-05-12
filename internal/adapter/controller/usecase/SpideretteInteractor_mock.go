//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockSpideretteInteractor スパイダレットインタラクターモック
type MockSpideretteInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSpideretteInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Deal モック
func (_m *MockSpideretteInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// MoveTableauToTableau モック
func (_m *MockSpideretteInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

// GiveUp モック
func (_m *MockSpideretteInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockSpideretteInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// AutoComplete モック
func (_m *MockSpideretteInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSpideretteInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Undo モック
func (_m *MockSpideretteInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// UndoN モック
func (_m *MockSpideretteInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSpideretteInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

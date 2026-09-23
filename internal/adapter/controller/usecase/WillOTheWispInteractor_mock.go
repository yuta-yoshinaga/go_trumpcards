//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockWillOTheWispInteractor ウィル・オ・ザ・ウィスプインタラクターモック
type MockWillOTheWispInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockWillOTheWispInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Deal モック
func (_m *MockWillOTheWispInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// MoveTableauToTableau モック
func (_m *MockWillOTheWispInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

// GiveUp モック
func (_m *MockWillOTheWispInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockWillOTheWispInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// AutoComplete モック
func (_m *MockWillOTheWispInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockWillOTheWispInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Undo モック
func (_m *MockWillOTheWispInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// UndoN モック
func (_m *MockWillOTheWispInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockWillOTheWispInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

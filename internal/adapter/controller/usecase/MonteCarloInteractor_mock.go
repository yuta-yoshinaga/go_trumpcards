//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockMonteCarloInteractor はモンテカルロ・ソリティアインタラクターのモック。
type MockMonteCarloInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockMonteCarloInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Remove モック
func (_m *MockMonteCarloInteractor) Remove(r1, c1, r2, c2 int) string {
	ret := _m.Called(r1, c1, r2, c2)
	return ret.Get(0).(string)
}

// Deal モック
func (_m *MockMonteCarloInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Undo モック
func (_m *MockMonteCarloInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GiveUp モック
func (_m *MockMonteCarloInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockMonteCarloInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockMonteCarloInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMonteCarloInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

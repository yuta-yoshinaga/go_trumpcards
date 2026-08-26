//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockFourteenOutInteractor はフォーティーンアウト・ソリティアインタラクターのモック。
type MockFourteenOutInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockFourteenOutInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Remove モック
func (_m *MockFourteenOutInteractor) Remove(c1, c2 int) string {
	ret := _m.Called(c1, c2)
	return ret.Get(0).(string)
}

// Undo モック
func (_m *MockFourteenOutInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GiveUp モック
func (_m *MockFourteenOutInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockFourteenOutInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockFourteenOutInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockFourteenOutInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

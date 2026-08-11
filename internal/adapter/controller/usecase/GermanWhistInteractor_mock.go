//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockGermanWhistInteractor ジャーマンホイストインタラクターモック
type MockGermanWhistInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockGermanWhistInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockGermanWhistInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockGermanWhistInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// GiveUp モック
func (_m *MockGermanWhistInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockGermanWhistInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockGermanWhistInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

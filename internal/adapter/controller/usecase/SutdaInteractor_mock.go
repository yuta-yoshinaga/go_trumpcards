//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSutdaInteractor はソッタのインタラクターモック。
type MockSutdaInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockSutdaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockSutdaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSutdaInteractor) ResetWithConfig(config domain.SutdaConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockSutdaInteractor) Action(action string) string {
	ret := _m.Called(action)
	return ret.Get(0).(string)
}

// NextHand モック
func (_m *MockSutdaInteractor) NextHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSutdaInteractor) GetConfig() domain.SutdaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SutdaConfig)
}

// Hint モック
func (_m *MockSutdaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSutdaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

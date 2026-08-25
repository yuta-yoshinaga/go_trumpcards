//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCometInteractor はコメットのインタラクターモック。
type MockCometInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockCometInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockCometInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCometInteractor) ResetWithConfig(config domain.CometConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCometInteractor) Play(handIdx int) string {
	ret := _m.Called(handIdx)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockCometInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCometInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCometInteractor) GetConfig() domain.CometConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CometConfig)
}

// Hint モック
func (_m *MockCometInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCometInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

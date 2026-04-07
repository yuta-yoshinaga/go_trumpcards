//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpeedInteractor スピードインタラクターモック
type MockSpeedInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSpeedInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSpeedInteractor) ResetWithConfig(cfg domain.SpeedConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSpeedInteractor) Play(cardIndex, pileIndex int) string {
	ret := _m.Called(cardIndex, pileIndex)
	return ret.Get(0).(string)
}

// Flip モック
func (_m *MockSpeedInteractor) Flip() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockSpeedInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSpeedInteractor) GetConfig() domain.SpeedConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SpeedConfig)
}

// ActionLog モック
func (_m *MockSpeedInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSpeedInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

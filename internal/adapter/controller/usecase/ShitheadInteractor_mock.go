//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShitheadInteractor シットヘッドインタラクターモック
type MockShitheadInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockShitheadInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockShitheadInteractor) Play(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockShitheadInteractor) GetConfig() domain.ShitheadConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ShitheadConfig)
}

// ResetWithConfig モック
func (_m *MockShitheadInteractor) ResetWithConfig(config domain.ShitheadConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockShitheadInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockShitheadInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

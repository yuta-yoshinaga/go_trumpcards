//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWarInteractor 戦争インタラクターモック
type MockWarInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockWarInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockWarInteractor) ResetWithConfig(cfg domain.WarConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Step モック
func (_m *MockWarInteractor) Step() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// AutoPlay モック
func (_m *MockWarInteractor) AutoPlay() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockWarInteractor) GetConfig() domain.WarConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.WarConfig)
}

// ActionLog モック
func (_m *MockWarInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockWarInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

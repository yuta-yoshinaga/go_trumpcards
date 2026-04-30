//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEgyptianRatscrewInteractor エジプシャン・ラットスクリューインタラクターモック
type MockEgyptianRatscrewInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockEgyptianRatscrewInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockEgyptianRatscrewInteractor) ResetWithConfig(cfg domain.EgyptianRatscrewConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Step モック
func (_m *MockEgyptianRatscrewInteractor) Step() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Slap モック
func (_m *MockEgyptianRatscrewInteractor) Slap(playerIdx int) string {
	ret := _m.Called(playerIdx)
	return ret.Get(0).(string)
}

// Tick モック
func (_m *MockEgyptianRatscrewInteractor) Tick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockEgyptianRatscrewInteractor) GetConfig() domain.EgyptianRatscrewConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.EgyptianRatscrewConfig)
}

// ActionLog モック
func (_m *MockEgyptianRatscrewInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockEgyptianRatscrewInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

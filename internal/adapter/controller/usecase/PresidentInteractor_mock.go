//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPresidentInteractor プレジデントインタラクターモック
type MockPresidentInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPresidentInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockPresidentInteractor) Play(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

func (_m *MockPresidentInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPresidentInteractor) GetConfig() domain.PresidentConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PresidentConfig)
}

// ResetWithConfig モック
func (_m *MockPresidentInteractor) ResetWithConfig(config domain.PresidentConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPresidentInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPresidentInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

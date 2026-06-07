//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTichuInteractor ティチューインタラクターモック
type MockTichuInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockTichuInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockTichuInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Declare モック
func (_m *MockTichuInteractor) Declare(declType int) string {
	ret := _m.Called(declType)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTichuInteractor) Play(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTichuInteractor) ResetWithConfig(config domain.TichuConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTichuInteractor) GetConfig() domain.TichuConfig {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.TichuConfig); ok {
		return val
	}
	return domain.TichuConfig{}
}

// ActionLog モック
func (_m *MockTichuInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockZhengInteractor モック
type MockZhengInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockZhengInteractor) Reset() string {
	return _m.Called().String(0)
}

// Play モック
func (_m *MockZhengInteractor) Play(indices []int) string {
	return _m.Called(indices).String(0)
}

// ResetWithConfig モック
func (_m *MockZhengInteractor) ResetWithConfig(cfg domain.ZhengConfig) string {
	return _m.Called(cfg).String(0)
}

// GetConfig モック
func (_m *MockZhengInteractor) GetConfig() domain.ZhengConfig {
	return _m.Called().Get(0).(domain.ZhengConfig)
}

// ActionLog モック
func (_m *MockZhengInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockZhengInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

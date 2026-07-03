//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGutsInteractor はガッツ (Guts) のインタラクターモック。
type MockGutsInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockGutsInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockGutsInteractor) ResetWithConfig(cfg domain.GutsConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Declare モック
func (_m *MockGutsInteractor) Declare(stay bool) string {
	ret := _m.Called(stay)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockGutsInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockGutsInteractor) GetConfig() domain.GutsConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GutsConfig)
}

// Hint モック
func (_m *MockGutsInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockGutsInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGutsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

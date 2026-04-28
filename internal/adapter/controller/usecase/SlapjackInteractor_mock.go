//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSlapjackInteractor スラップジャックインタラクターモック
type MockSlapjackInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSlapjackInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSlapjackInteractor) ResetWithConfig(cfg domain.SlapjackConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Step モック
func (_m *MockSlapjackInteractor) Step() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Slap モック
func (_m *MockSlapjackInteractor) Slap(playerIdx int) string {
	ret := _m.Called(playerIdx)
	return ret.Get(0).(string)
}

// Tick モック
func (_m *MockSlapjackInteractor) Tick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSlapjackInteractor) GetConfig() domain.SlapjackConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SlapjackConfig)
}

// ActionLog モック
func (_m *MockSlapjackInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSlapjackInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

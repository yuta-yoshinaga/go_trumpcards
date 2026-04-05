//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPigsTailInteractor ぶたのしっぽインタラクターモック
type MockPigsTailInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPigsTailInteractor) Reset(config domain.PigsTailConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPigsTailInteractor) GetConfig() domain.PigsTailConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PigsTailConfig)
}

// Action モック
func (_m *MockPigsTailInteractor) Action(actionIdx int) string {
	ret := _m.Called(actionIdx)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPigsTailInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

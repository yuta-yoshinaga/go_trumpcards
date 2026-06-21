//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPishtiInteractor は Pişti インタラクターモック。
type MockPishtiInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPishtiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockPishtiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockPishtiInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPishtiInteractor) GetConfig() domain.PishtiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PishtiConfig)
}

// ResetWithConfig モック
func (_m *MockPishtiInteractor) ResetWithConfig(config domain.PishtiConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPishtiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPishtiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

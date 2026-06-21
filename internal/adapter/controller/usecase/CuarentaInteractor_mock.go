//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCuarentaInteractor クアレンタインタラクターモック。
type MockCuarentaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCuarentaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCuarentaInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCuarentaInteractor) Play(handIdx int) string {
	ret := _m.Called(handIdx)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCuarentaInteractor) GetConfig() domain.CuarentaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CuarentaConfig)
}

// ResetWithConfig モック
func (_m *MockCuarentaInteractor) ResetWithConfig(config domain.CuarentaConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCuarentaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCuarentaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

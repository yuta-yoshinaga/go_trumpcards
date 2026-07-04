//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPrimeroInteractor はプリメロ (Primero) のインタラクターモック。
type MockPrimeroInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPrimeroInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockPrimeroInteractor) ResetWithConfig(cfg domain.PrimeroConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bet モック
func (_m *MockPrimeroInteractor) Bet(action string) string {
	ret := _m.Called(action)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockPrimeroInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPrimeroInteractor) GetConfig() domain.PrimeroConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PrimeroConfig)
}

// Hint モック
func (_m *MockPrimeroInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPrimeroInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPrimeroInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

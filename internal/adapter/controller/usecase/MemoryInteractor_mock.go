//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMemoryInteractor 神経衰弱インタラクターモック
type MockMemoryInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockMemoryInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockMemoryInteractor) ResetWithConfig(cfg domain.MemoryConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Flip モック
func (_m *MockMemoryInteractor) Flip(pos int) string {
	ret := _m.Called(pos)
	return ret.Get(0).(string)
}

// Next モック
func (_m *MockMemoryInteractor) Next() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockMemoryInteractor) GetConfig() domain.MemoryConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MemoryConfig)
}

// ActionLog モック
func (_m *MockMemoryInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMemoryInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHachiHachiInteractor は八八 (Hachi-Hachi) のインタラクターモック。
type MockHachiHachiInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockHachiHachiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockHachiHachiInteractor) ResetWithConfig(cfg domain.HachiHachiConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockHachiHachiInteractor) Play(handIdx, fieldIdx int) string {
	ret := _m.Called(handIdx, fieldIdx)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockHachiHachiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockHachiHachiInteractor) GetConfig() domain.HachiHachiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.HachiHachiConfig)
}

// Hint モック
func (_m *MockHachiHachiInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockHachiHachiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockHachiHachiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

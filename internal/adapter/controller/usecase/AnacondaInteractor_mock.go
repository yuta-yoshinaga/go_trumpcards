//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAnacondaInteractor はアナコンダ (Anaconda) のインタラクターモック。
type MockAnacondaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockAnacondaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockAnacondaInteractor) ResetWithConfig(cfg domain.AnacondaConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockAnacondaInteractor) Pass(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Keep モック
func (_m *MockAnacondaInteractor) Keep(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Call モック
func (_m *MockAnacondaInteractor) Call() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Raise モック
func (_m *MockAnacondaInteractor) Raise() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Fold モック
func (_m *MockAnacondaInteractor) Fold() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockAnacondaInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockAnacondaInteractor) GetConfig() domain.AnacondaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.AnacondaConfig)
}

// Hint モック
func (_m *MockAnacondaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockAnacondaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockAnacondaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

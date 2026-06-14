//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMusInteractor ムスのインタラクターモック
type MockMusInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockMusInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockMusInteractor) ResetWithConfig(cfg domain.MusConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Mus モック
func (_m *MockMusInteractor) Mus(mus bool) string {
	ret := _m.Called(mus)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockMusInteractor) Discard(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Bet モック
func (_m *MockMusInteractor) Bet(action, amount int) string {
	ret := _m.Called(action, amount)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockMusInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockMusInteractor) GetConfig() domain.MusConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MusConfig)
}

// Hint モック
func (_m *MockMusInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockMusInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMusInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHorseInteractor は H.O.R.S.E. のインタラクターモック。
type MockHorseInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockHorseInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockHorseInteractor) ResetWithConfig(cfg domain.HorseConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockHorseInteractor) Action(action, amount, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// NextHand モック
func (_m *MockHorseInteractor) NextHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockHorseInteractor) GetConfig() domain.HorseConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.HorseConfig)
}

// Hint モック
func (_m *MockHorseInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockHorseInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockHorseInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDilotiInteractor はディロティのインタラクターモック。
type MockDilotiInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockDilotiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockDilotiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockDilotiInteractor) ResetWithConfig(config domain.DilotiConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockDilotiInteractor) Play(handIdx int, action string, tableIdxs, declIdxs []int, declValue int) string {
	ret := _m.Called(handIdx, action, tableIdxs, declIdxs, declValue)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockDilotiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockDilotiInteractor) GetConfig() domain.DilotiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DilotiConfig)
}

// Hint モック
func (_m *MockDilotiInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockDilotiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

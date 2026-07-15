//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCassinoInteractor カシノインタラクターモック。
type MockCassinoInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCassinoInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCassinoInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Take モック
func (_m *MockCassinoInteractor) Take(handIdx int, tableIdxs []int, buildIdxs []int) string {
	ret := _m.Called(handIdx, tableIdxs, buildIdxs)
	return ret.Get(0).(string)
}

// Build モック
func (_m *MockCassinoInteractor) Build(handIdx int, tableIdxs []int, declaredValue int) string {
	ret := _m.Called(handIdx, tableIdxs, declaredValue)
	return ret.Get(0).(string)
}

// Trail モック
func (_m *MockCassinoInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCassinoInteractor) Trail(handIdx int) string {
	ret := _m.Called(handIdx)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCassinoInteractor) GetConfig() domain.CassinoConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CassinoConfig)
}

// ResetWithConfig モック
func (_m *MockCassinoInteractor) ResetWithConfig(config domain.CassinoConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCassinoInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCassinoInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

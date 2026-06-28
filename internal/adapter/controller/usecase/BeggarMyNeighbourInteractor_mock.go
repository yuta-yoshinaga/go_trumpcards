//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBeggarMyNeighbourInteractor Beggar-My-Neighbour インタラクターモック
type MockBeggarMyNeighbourInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBeggarMyNeighbourInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBeggarMyNeighbourInteractor) ResetWithConfig(cfg domain.BeggarMyNeighbourConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Step モック
func (_m *MockBeggarMyNeighbourInteractor) Step() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// AutoPlay モック
func (_m *MockBeggarMyNeighbourInteractor) AutoPlay() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBeggarMyNeighbourInteractor) GetConfig() domain.BeggarMyNeighbourConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BeggarMyNeighbourConfig)
}

// ActionLog モック
func (_m *MockBeggarMyNeighbourInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBeggarMyNeighbourInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

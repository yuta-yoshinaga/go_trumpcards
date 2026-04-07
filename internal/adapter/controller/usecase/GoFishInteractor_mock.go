//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGoFishInteractor Go Fishインタラクターモック
type MockGoFishInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockGoFishInteractor) Reset(config domain.GoFishConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockGoFishInteractor) GetConfig() domain.GoFishConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GoFishConfig)
}

// Ask モック
func (_m *MockGoFishInteractor) Ask(targetIdx, rank int) string {
	ret := _m.Called(targetIdx, rank)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockGoFishInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGoFishInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

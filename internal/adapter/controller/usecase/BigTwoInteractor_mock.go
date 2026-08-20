//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBigTwoInteractor 大富豪(Big Two)のインタラクターモック
type MockBigTwoInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockBigTwoInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockBigTwoInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBigTwoInteractor) Play(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBigTwoInteractor) ResetWithConfig(config domain.BigTwoConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBigTwoInteractor) GetConfig() domain.BigTwoConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BigTwoConfig)
}

// ActionLog モック
func (_m *MockBigTwoInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

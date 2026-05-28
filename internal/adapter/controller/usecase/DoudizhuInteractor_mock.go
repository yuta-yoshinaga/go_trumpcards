//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDoudizhuInteractor 斗地主インタラクターモック
type MockDoudizhuInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockDoudizhuInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockDoudizhuInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockDoudizhuInteractor) Bid(value int) string {
	ret := _m.Called(value)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockDoudizhuInteractor) Play(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockDoudizhuInteractor) ResetWithConfig(config domain.DoudizhuConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockDoudizhuInteractor) GetConfig() domain.DoudizhuConfig {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.DoudizhuConfig); ok {
		return val
	}
	return domain.DoudizhuConfig{}
}

// ActionLog モック
func (_m *MockDoudizhuInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

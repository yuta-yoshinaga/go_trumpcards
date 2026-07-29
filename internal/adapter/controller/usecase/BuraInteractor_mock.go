//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBuraInteractor ブラ インタラクターモック
type MockBuraInteractor struct {
	mock.Mock
}

func (_m *MockBuraInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBuraInteractor) ResetWithConfig(cfg domain.BuraConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

func (_m *MockBuraInteractor) Play(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

func (_m *MockBuraInteractor) Claim() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBuraInteractor) Declare() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBuraInteractor) GetConfig() domain.BuraConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BuraConfig)
}

func (_m *MockBuraInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBuraInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBuraInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

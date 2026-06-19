//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCuckooInteractor モック
type MockCuckooInteractor struct {
	mock.Mock
}

func (_m *MockCuckooInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockCuckooInteractor) ResetWithConfig(cfg domain.CuckooConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockCuckooInteractor) Keep() string {
	return _m.Called().String(0)
}

func (_m *MockCuckooInteractor) Swap() string {
	return _m.Called().String(0)
}

func (_m *MockCuckooInteractor) Refuse() string {
	return _m.Called().String(0)
}

func (_m *MockCuckooInteractor) AcceptSwap() string {
	return _m.Called().String(0)
}

func (_m *MockCuckooInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockCuckooInteractor) GetConfig() domain.CuckooConfig {
	return _m.Called().Get(0).(domain.CuckooConfig)
}

func (_m *MockCuckooInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockCuckooInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

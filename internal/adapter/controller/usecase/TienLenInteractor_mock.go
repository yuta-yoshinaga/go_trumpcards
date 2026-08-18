//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTienLenInteractor モック
type MockTienLenInteractor struct {
	mock.Mock
}

func (_m *MockTienLenInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockTienLenInteractor) Play(indices []int) string {
	return _m.Called(indices).String(0)
}

func (_m *MockTienLenInteractor) ResetWithConfig(cfg domain.TienLenConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockTienLenInteractor) GetConfig() domain.TienLenConfig {
	return _m.Called().Get(0).(domain.TienLenConfig)
}

func (_m *MockTienLenInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockTienLenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// Hint モック
func (_m *MockTienLenInteractor) Hint() string {
	ret := _m.Called()
	return ret.String(0)
}

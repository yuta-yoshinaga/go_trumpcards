//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRamschInteractor mocks the Ramsch interactor.
type MockRamschInteractor struct {
	mock.Mock
}

// Reset mock.
func (_m *MockRamschInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig mock.
func (_m *MockRamschInteractor) ResetWithConfig(cfg domain.RamschConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Play mock.
func (_m *MockRamschInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextTrick mock.
func (_m *MockRamschInteractor) NextTrick() string { return _m.Called().Get(0).(string) }

// NextRound mock.
func (_m *MockRamschInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GetConfig mock.
func (_m *MockRamschInteractor) GetConfig() domain.RamschConfig {
	return _m.Called().Get(0).(domain.RamschConfig)
}

// Hint mock.
func (_m *MockRamschInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog mock.
func (_m *MockRamschInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot mock.
func (_m *MockRamschInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

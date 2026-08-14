//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPigInteractor ピッグインタラクターモック
type MockPigInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockPigInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockPigInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockPigInteractor) ResetWithConfig(cfg domain.PigConfig) string {
	return _m.Called(cfg).String(0)
}

// Pass モック
func (_m *MockPigInteractor) Pass(cardIndex int) string { return _m.Called(cardIndex).String(0) }

// Signal モック
func (_m *MockPigInteractor) Signal() string { return _m.Called().String(0) }

// NextRound モック
func (_m *MockPigInteractor) NextRound() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockPigInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockPigInteractor) GetConfig() domain.PigConfig {
	return _m.Called().Get(0).(domain.PigConfig)
}

// Hint モック
func (_m *MockPigInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockPigInteractor) ActionLog() string { return _m.Called().String(0) }

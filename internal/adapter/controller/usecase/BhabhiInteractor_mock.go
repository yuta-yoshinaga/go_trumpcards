//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBhabhiInteractor バービーインタラクターモック
type MockBhabhiInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockBhabhiInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockBhabhiInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockBhabhiInteractor) ResetWithConfig(cfg domain.BhabhiConfig) string {
	return _m.Called(cfg).String(0)
}

// Play モック
func (_m *MockBhabhiInteractor) Play(cardIndex int) string { return _m.Called(cardIndex).String(0) }

// GiveUp モック
func (_m *MockBhabhiInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockBhabhiInteractor) GetConfig() domain.BhabhiConfig {
	return _m.Called().Get(0).(domain.BhabhiConfig)
}

// Hint モック
func (_m *MockBhabhiInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockBhabhiInteractor) ActionLog() string { return _m.Called().String(0) }

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPasurInteractor パスールインタラクターモック
type MockPasurInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockPasurInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockPasurInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockPasurInteractor) ResetWithConfig(cfg domain.PasurConfig) string {
	return _m.Called(cfg).String(0)
}

// Play モック
func (_m *MockPasurInteractor) Play(cardIndex int, tableIndices []int) string {
	return _m.Called(cardIndex, tableIndices).String(0)
}

// GiveUp モック
func (_m *MockPasurInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockPasurInteractor) GetConfig() domain.PasurConfig {
	return _m.Called().Get(0).(domain.PasurConfig)
}

// Hint モック
func (_m *MockPasurInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockPasurInteractor) ActionLog() string { return _m.Called().String(0) }

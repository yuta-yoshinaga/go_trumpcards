//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSnapInteractor スナップインタラクターモック
type MockSnapInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockSnapInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockSnapInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockSnapInteractor) ResetWithConfig(cfg domain.SnapConfig) string {
	return _m.Called(cfg).String(0)
}

// Step モック
func (_m *MockSnapInteractor) Step() string { return _m.Called().String(0) }

// Snap モック
func (_m *MockSnapInteractor) Snap() string { return _m.Called().String(0) }

// Tick モック
func (_m *MockSnapInteractor) Tick() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockSnapInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockSnapInteractor) GetConfig() domain.SnapConfig {
	return _m.Called().Get(0).(domain.SnapConfig)
}

// Hint モック
func (_m *MockSnapInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockSnapInteractor) ActionLog() string { return _m.Called().String(0) }

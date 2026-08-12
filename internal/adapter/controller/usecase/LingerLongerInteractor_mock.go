//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLingerLongerInteractor リンガーロンガーインタラクターモック
type MockLingerLongerInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockLingerLongerInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockLingerLongerInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockLingerLongerInteractor) ResetWithConfig(cfg domain.LingerLongerConfig) string {
	return _m.Called(cfg).String(0)
}

// Play モック
func (_m *MockLingerLongerInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

// GiveUp モック
func (_m *MockLingerLongerInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockLingerLongerInteractor) GetConfig() domain.LingerLongerConfig {
	return _m.Called().Get(0).(domain.LingerLongerConfig)
}

// Hint モック
func (_m *MockLingerLongerInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockLingerLongerInteractor) ActionLog() string { return _m.Called().String(0) }

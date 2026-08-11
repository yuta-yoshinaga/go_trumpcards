//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMinibridgeInteractor ミニブリッジインタラクターモック
type MockMinibridgeInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockMinibridgeInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockMinibridgeInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockMinibridgeInteractor) ResetWithConfig(cfg domain.MinibridgeConfig) string {
	return _m.Called(cfg).String(0)
}

// Contract モック
func (_m *MockMinibridgeInteractor) Contract(level, suit int) string {
	return _m.Called(level, suit).String(0)
}

// Play モック
func (_m *MockMinibridgeInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

// NextRound モック
func (_m *MockMinibridgeInteractor) NextRound() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockMinibridgeInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockMinibridgeInteractor) GetConfig() domain.MinibridgeConfig {
	return _m.Called().Get(0).(domain.MinibridgeConfig)
}

// Hint モック
func (_m *MockMinibridgeInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockMinibridgeInteractor) ActionLog() string { return _m.Called().String(0) }

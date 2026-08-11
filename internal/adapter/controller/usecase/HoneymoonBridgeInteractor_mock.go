//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHoneymoonBridgeInteractor ハネムーンブリッジインタラクターモック
type MockHoneymoonBridgeInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockHoneymoonBridgeInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockHoneymoonBridgeInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockHoneymoonBridgeInteractor) ResetWithConfig(cfg domain.HoneymoonBridgeConfig) string {
	return _m.Called(cfg).String(0)
}

// Bid モック
func (_m *MockHoneymoonBridgeInteractor) Bid(level, suit int) string {
	return _m.Called(level, suit).String(0)
}

// Pass モック
func (_m *MockHoneymoonBridgeInteractor) Pass() string { return _m.Called().String(0) }

// Play モック
func (_m *MockHoneymoonBridgeInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

// NextRound モック
func (_m *MockHoneymoonBridgeInteractor) NextRound() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockHoneymoonBridgeInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockHoneymoonBridgeInteractor) GetConfig() domain.HoneymoonBridgeConfig {
	return _m.Called().Get(0).(domain.HoneymoonBridgeConfig)
}

// Hint モック
func (_m *MockHoneymoonBridgeInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockHoneymoonBridgeInteractor) ActionLog() string { return _m.Called().String(0) }

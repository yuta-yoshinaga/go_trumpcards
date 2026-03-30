//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBridgeInteractor ブリッジインタラクターモック
type MockBridgeInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBridgeInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBridgeInteractor) ResetWithConfig(cfg domain.BridgeConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockBridgeInteractor) Bid(bidType int, level int, suit int) string {
	ret := _m.Called(bidType, level, suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBridgeInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockBridgeInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBridgeInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBridgeInteractor) GetConfig() domain.BridgeConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BridgeConfig)
}

// Hint モック
func (_m *MockBridgeInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBridgeInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

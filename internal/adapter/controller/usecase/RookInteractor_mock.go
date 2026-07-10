//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRookInteractor ルーク(Rook)インタラクターモック
type MockRookInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockRookInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockRookInteractor) ResetWithConfig(cfg domain.RookConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockRookInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockRookInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ExchangeNest モック
func (_m *MockRookInteractor) ExchangeNest(discardIndices []int, trumpColor int) string {
	ret := _m.Called(discardIndices, trumpColor)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockRookInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockRookInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockRookInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockRookInteractor) GetConfig() domain.RookConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.RookConfig)
}

// Hint モック
func (_m *MockRookInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockRookInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockRookInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

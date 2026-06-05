//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBidWhistInteractor Bid Whist インタラクターモック
type MockBidWhistInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBidWhistInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBidWhistInteractor) ResetWithConfig(cfg domain.BidWhistConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockBidWhistInteractor) Bid(tricks, direction int) string {
	ret := _m.Called(tricks, direction)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockBidWhistInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// DeclareTrump モック
func (_m *MockBidWhistInteractor) DeclareTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// ExchangeKitty モック
func (_m *MockBidWhistInteractor) ExchangeKitty(discardIndices []int) string {
	ret := _m.Called(discardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBidWhistInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockBidWhistInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBidWhistInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBidWhistInteractor) GetConfig() domain.BidWhistConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BidWhistConfig)
}

// Hint モック
func (_m *MockBidWhistInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBidWhistInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBidWhistInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

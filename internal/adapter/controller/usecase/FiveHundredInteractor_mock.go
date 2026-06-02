//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFiveHundredInteractor 500インタラクターモック
type MockFiveHundredInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockFiveHundredInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockFiveHundredInteractor) ResetWithConfig(cfg domain.FiveHundredConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockFiveHundredInteractor) Bid(kind domain.FiveHundredContractKind, tricks, suit int) string {
	ret := _m.Called(kind, tricks, suit)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockFiveHundredInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ExchangeKitty モック
func (_m *MockFiveHundredInteractor) ExchangeKitty(discardIndices []int) string {
	ret := _m.Called(discardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockFiveHundredInteractor) Play(cardIndex, jokerSuit int) string {
	ret := _m.Called(cardIndex, jokerSuit)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockFiveHundredInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockFiveHundredInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockFiveHundredInteractor) GetConfig() domain.FiveHundredConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.FiveHundredConfig)
}

// Hint モック
func (_m *MockFiveHundredInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockFiveHundredInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockFiveHundredInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPinochleInteractor ピノクルインタラクターモック
type MockPinochleInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPinochleInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockPinochleInteractor) ResetWithConfig(cfg domain.PinochleConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockPinochleInteractor) Bid(amount int) string {
	ret := _m.Called(amount)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockPinochleInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// CallTrump モック
func (_m *MockPinochleInteractor) CallTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// ConfirmMelds モック
func (_m *MockPinochleInteractor) ConfirmMelds() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockPinochleInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockPinochleInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockPinochleInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPinochleInteractor) GetConfig() domain.PinochleConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PinochleConfig)
}

// Hint モック
func (_m *MockPinochleInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPinochleInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

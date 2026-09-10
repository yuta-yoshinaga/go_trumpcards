//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTappTarockInteractor タップ・タロックのインタラクターモック。
type MockTappTarockInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTappTarockInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTappTarockInteractor) ResetWithConfig(cfg domain.TappTarockConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockTappTarockInteractor) Bid(bid domain.TappTarockBid) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockTappTarockInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockTappTarockInteractor) Discard(cardIndices []int) string {
	ret := _m.Called(cardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTappTarockInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockTappTarockInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTappTarockInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTappTarockInteractor) GetConfig() domain.TappTarockConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TappTarockConfig)
}

// Hint モック
func (_m *MockTappTarockInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTappTarockInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTappTarockInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

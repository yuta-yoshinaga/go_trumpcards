//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrogguInteractor トロッグのインタラクターモック。
type MockTrogguInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTrogguInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTrogguInteractor) ResetWithConfig(cfg domain.TrogguConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockTrogguInteractor) Bid(bid domain.TrogguBid) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockTrogguInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTrogguInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockTrogguInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTrogguInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTrogguInteractor) GetConfig() domain.TrogguConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TrogguConfig)
}

// Hint モック
func (_m *MockTrogguInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTrogguInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTrogguInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

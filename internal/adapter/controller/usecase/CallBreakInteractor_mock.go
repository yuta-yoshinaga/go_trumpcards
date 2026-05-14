//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCallBreakInteractor Call Break インタラクターモック
type MockCallBreakInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCallBreakInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCallBreakInteractor) ResetWithConfig(cfg domain.CallBreakConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockCallBreakInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCallBreakInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockCallBreakInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCallBreakInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCallBreakInteractor) GetConfig() domain.CallBreakConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CallBreakConfig)
}

// Hint モック
func (_m *MockCallBreakInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCallBreakInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCallBreakInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

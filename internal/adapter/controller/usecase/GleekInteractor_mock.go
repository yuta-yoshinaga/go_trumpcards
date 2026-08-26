//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGleekInteractor グリーク (Gleek) のインタラクターモック
type MockGleekInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockGleekInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockGleekInteractor) ResetWithConfig(cfg domain.GleekConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockGleekInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockGleekInteractor) Discard(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockGleekInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockGleekInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockGleekInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockGleekInteractor) GetConfig() domain.GleekConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GleekConfig)
}

// Hint モック
func (_m *MockGleekInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockGleekInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGleekInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

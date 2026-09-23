//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBatakInteractor Batak インタラクターモック
type MockBatakInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBatakInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBatakInteractor) ResetWithConfig(cfg domain.BatakConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockBatakInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBatakInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockBatakInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBatakInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBatakInteractor) GetConfig() domain.BatakConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BatakConfig)
}

// Hint モック
func (_m *MockBatakInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBatakInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBatakInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

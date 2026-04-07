//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHeartsInteractor ハーツインタラクターモック
type MockHeartsInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockHeartsInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockHeartsInteractor) ResetWithConfig(cfg domain.HeartsConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockHeartsInteractor) Pass(cardIndices []int) string {
	ret := _m.Called(cardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockHeartsInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockHeartsInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockHeartsInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockHeartsInteractor) GetConfig() domain.HeartsConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.HeartsConfig)
}

// Hint モック
func (_m *MockHeartsInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockHeartsInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockHeartsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

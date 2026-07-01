//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLooInteractor はルー (Loo) のインタラクターモック。
type MockLooInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockLooInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockLooInteractor) ResetWithConfig(cfg domain.LooConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Decide モック
func (_m *MockLooInteractor) Decide(play bool) string {
	ret := _m.Called(play)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockLooInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockLooInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockLooInteractor) GetConfig() domain.LooConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.LooConfig)
}

// Hint モック
func (_m *MockLooInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockLooInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockLooInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGoStopInteractor はゴーストップ (Go-Stop) のインタラクターモック。
type MockGoStopInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockGoStopInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockGoStopInteractor) ResetWithConfig(cfg domain.GoStopConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockGoStopInteractor) Play(handIdx, fieldIdx int) string {
	ret := _m.Called(handIdx, fieldIdx)
	return ret.Get(0).(string)
}

// Decide モック
func (_m *MockGoStopInteractor) Decide(goDecision bool) string {
	ret := _m.Called(goDecision)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockGoStopInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockGoStopInteractor) GetConfig() domain.GoStopConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GoStopConfig)
}

// Hint モック
func (_m *MockGoStopInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockGoStopInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGoStopInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

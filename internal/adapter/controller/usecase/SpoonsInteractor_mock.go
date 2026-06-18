//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpoonsInteractor はスプーンのインタラクターモック。
type MockSpoonsInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSpoonsInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSpoonsInteractor) ResetWithConfig(cfg domain.SpoonsConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockSpoonsInteractor) Pass(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// Grab モック
func (_m *MockSpoonsInteractor) Grab() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockSpoonsInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSpoonsInteractor) GetConfig() domain.SpoonsConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SpoonsConfig)
}

// ActionLog モック
func (_m *MockSpoonsInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSpoonsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

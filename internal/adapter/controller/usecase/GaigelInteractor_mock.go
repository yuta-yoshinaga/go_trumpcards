//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGaigelInteractor ガイゲルインタラクターモック
type MockGaigelInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockGaigelInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockGaigelInteractor) ResetWithConfig(cfg domain.GaigelConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockGaigelInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// DeclareMarriage モック
func (_m *MockGaigelInteractor) DeclareMarriage(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockGaigelInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockGaigelInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockGaigelInteractor) GetConfig() domain.GaigelConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GaigelConfig)
}

// Hint モック
func (_m *MockGaigelInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockGaigelInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGaigelInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

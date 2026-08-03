//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMinchiateInteractor ミンキアーテのインタラクターモック
type MockMinchiateInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockMinchiateInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockMinchiateInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockMinchiateInteractor) ResetWithConfig(cfg domain.MinchiateConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockMinchiateInteractor) Discard(cardIndices []int) string {
	ret := _m.Called(cardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockMinchiateInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockMinchiateInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockMinchiateInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockMinchiateInteractor) GetConfig() domain.MinchiateConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MinchiateConfig)
}

// Hint モック
func (_m *MockMinchiateInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockMinchiateInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

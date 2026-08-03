//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTarocchiniInteractor タロッキーニのインタラクターモック
type MockTarocchiniInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockTarocchiniInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockTarocchiniInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTarocchiniInteractor) ResetWithConfig(cfg domain.TarocchiniConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockTarocchiniInteractor) Discard(cardIndices []int) string {
	ret := _m.Called(cardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTarocchiniInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockTarocchiniInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTarocchiniInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTarocchiniInteractor) GetConfig() domain.TarocchiniConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TarocchiniConfig)
}

// Hint モック
func (_m *MockTarocchiniInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTarocchiniInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

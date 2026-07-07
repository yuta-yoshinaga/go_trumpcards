//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockScartoInteractor スカルト (Scarto) のインタラクターモック
type MockScartoInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockScartoInteractor) Reset() string {
	return _m.Called().Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockScartoInteractor) ResetWithConfig(cfg domain.ScartoConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Discard モック
func (_m *MockScartoInteractor) Discard(cardIndices []int) string {
	return _m.Called(cardIndices).Get(0).(string)
}

// Play モック
func (_m *MockScartoInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextTrick モック
func (_m *MockScartoInteractor) NextTrick() string {
	return _m.Called().Get(0).(string)
}

// NextRound モック
func (_m *MockScartoInteractor) NextRound() string {
	return _m.Called().Get(0).(string)
}

// GetConfig モック
func (_m *MockScartoInteractor) GetConfig() domain.ScartoConfig {
	return _m.Called().Get(0).(domain.ScartoConfig)
}

// Hint モック
func (_m *MockScartoInteractor) Hint() string {
	return _m.Called().Get(0).(string)
}

// ActionLog モック
func (_m *MockScartoInteractor) ActionLog() string {
	return _m.Called().Get(0).(string)
}

// Snapshot モック
func (_m *MockScartoInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

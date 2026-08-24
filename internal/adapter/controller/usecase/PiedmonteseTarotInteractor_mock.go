//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPiedmonteseTarotInteractor はピエモンテ・タロッコのインタラクターモック。
type MockPiedmonteseTarotInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPiedmonteseTarotInteractor) Reset() string {
	return _m.Called().Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockPiedmonteseTarotInteractor) ResetWithConfig(cfg domain.PiedmonteseTarotConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Discard モック
func (_m *MockPiedmonteseTarotInteractor) Discard(cardIndices []int) string {
	return _m.Called(cardIndices).Get(0).(string)
}

// Play モック
func (_m *MockPiedmonteseTarotInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextTrick モック
func (_m *MockPiedmonteseTarotInteractor) NextTrick() string {
	return _m.Called().Get(0).(string)
}

// NextRound モック
func (_m *MockPiedmonteseTarotInteractor) NextRound() string {
	return _m.Called().Get(0).(string)
}

// GetConfig モック
func (_m *MockPiedmonteseTarotInteractor) GetConfig() domain.PiedmonteseTarotConfig {
	return _m.Called().Get(0).(domain.PiedmonteseTarotConfig)
}

// Hint モック
func (_m *MockPiedmonteseTarotInteractor) Hint() string {
	return _m.Called().Get(0).(string)
}

// ActionLog モック
func (_m *MockPiedmonteseTarotInteractor) ActionLog() string {
	return _m.Called().Get(0).(string)
}

// Snapshot モック
func (_m *MockPiedmonteseTarotInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

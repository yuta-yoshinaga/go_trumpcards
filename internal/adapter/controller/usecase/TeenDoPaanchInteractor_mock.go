//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTeenDoPaanchInteractor 3-2-5 インタラクターモック
type MockTeenDoPaanchInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockTeenDoPaanchInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockTeenDoPaanchInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockTeenDoPaanchInteractor) ResetWithConfig(cfg domain.TeenDoPaanchConfig) string {
	return _m.Called(cfg).String(0)
}

// DeclareTrump モック
func (_m *MockTeenDoPaanchInteractor) DeclareTrump(suit int) string {
	return _m.Called(suit).String(0)
}

// Play モック
func (_m *MockTeenDoPaanchInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

// NextRound モック
func (_m *MockTeenDoPaanchInteractor) NextRound() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockTeenDoPaanchInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockTeenDoPaanchInteractor) GetConfig() domain.TeenDoPaanchConfig {
	return _m.Called().Get(0).(domain.TeenDoPaanchConfig)
}

// Hint モック
func (_m *MockTeenDoPaanchInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockTeenDoPaanchInteractor) ActionLog() string { return _m.Called().String(0) }

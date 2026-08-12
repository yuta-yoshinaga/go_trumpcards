//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCucumberInteractor キューカンバーインタラクターモック
type MockCucumberInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockCucumberInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockCucumberInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockCucumberInteractor) ResetWithConfig(cfg domain.CucumberConfig) string {
	return _m.Called(cfg).String(0)
}

// Play モック
func (_m *MockCucumberInteractor) Play(cardIndex int) string { return _m.Called(cardIndex).String(0) }

// NextRound モック
func (_m *MockCucumberInteractor) NextRound() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockCucumberInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockCucumberInteractor) GetConfig() domain.CucumberConfig {
	return _m.Called().Get(0).(domain.CucumberConfig)
}

// Hint モック
func (_m *MockCucumberInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockCucumberInteractor) ActionLog() string { return _m.Called().String(0) }

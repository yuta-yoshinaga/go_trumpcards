//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockJulepeInteractor フレペインタラクターモック
type MockJulepeInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockJulepeInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockJulepeInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockJulepeInteractor) ResetWithConfig(cfg domain.JulepeConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Play モック
func (_m *MockJulepeInteractor) Play() string { return _m.Called().Get(0).(string) }

// Pass モック
func (_m *MockJulepeInteractor) Pass() string { return _m.Called().Get(0).(string) }

// PlayCard モック
func (_m *MockJulepeInteractor) PlayCard(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockJulepeInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockJulepeInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockJulepeInteractor) GetConfig() domain.JulepeConfig {
	return _m.Called().Get(0).(domain.JulepeConfig)
}

// Hint モック
func (_m *MockJulepeInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockJulepeInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

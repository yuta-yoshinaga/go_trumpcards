//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRamsInteractor ラムスインタラクターモック
type MockRamsInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockRamsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockRamsInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockRamsInteractor) ResetWithConfig(cfg domain.RamsConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Play モック
func (_m *MockRamsInteractor) Play() string { return _m.Called().Get(0).(string) }

// Pass モック
func (_m *MockRamsInteractor) Pass() string { return _m.Called().Get(0).(string) }

// PlayCard モック
func (_m *MockRamsInteractor) PlayCard(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockRamsInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockRamsInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockRamsInteractor) GetConfig() domain.RamsConfig {
	return _m.Called().Get(0).(domain.RamsConfig)
}

// Hint モック
func (_m *MockRamsInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockRamsInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

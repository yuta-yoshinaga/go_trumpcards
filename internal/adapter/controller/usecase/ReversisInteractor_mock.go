//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockReversisInteractor レヴェルシインタラクターモック
type MockReversisInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockReversisInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockReversisInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockReversisInteractor) ResetWithConfig(cfg domain.ReversisConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Play モック
func (_m *MockReversisInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockReversisInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockReversisInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockReversisInteractor) GetConfig() domain.ReversisConfig {
	return _m.Called().Get(0).(domain.ReversisConfig)
}

// Hint モック
func (_m *MockReversisInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockReversisInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

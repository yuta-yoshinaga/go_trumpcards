//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSlobberhannesInteractor スロバーハンネスインタラクターモック
type MockSlobberhannesInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockSlobberhannesInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockSlobberhannesInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockSlobberhannesInteractor) ResetWithConfig(cfg domain.SlobberhannesConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Play モック
func (_m *MockSlobberhannesInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockSlobberhannesInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockSlobberhannesInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockSlobberhannesInteractor) GetConfig() domain.SlobberhannesConfig {
	return _m.Called().Get(0).(domain.SlobberhannesConfig)
}

// Hint モック
func (_m *MockSlobberhannesInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockSlobberhannesInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

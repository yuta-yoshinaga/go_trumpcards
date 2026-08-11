//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMendikotInteractor メンディコットインタラクターモック
type MockMendikotInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockMendikotInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockMendikotInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockMendikotInteractor) ResetWithConfig(cfg domain.MendikotConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Play モック
func (_m *MockMendikotInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextHand モック
func (_m *MockMendikotInteractor) NextHand() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockMendikotInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockMendikotInteractor) GetConfig() domain.MendikotConfig {
	return _m.Called().Get(0).(domain.MendikotConfig)
}

// Hint モック
func (_m *MockMendikotInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockMendikotInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

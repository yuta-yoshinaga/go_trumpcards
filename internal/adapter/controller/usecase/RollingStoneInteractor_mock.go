//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRollingStoneInteractor ローリングストーンインタラクターモック
type MockRollingStoneInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockRollingStoneInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockRollingStoneInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockRollingStoneInteractor) ResetWithConfig(cfg domain.RollingStoneConfig) string {
	return _m.Called(cfg).String(0)
}

// Play モック
func (_m *MockRollingStoneInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

// PickUp モック
func (_m *MockRollingStoneInteractor) PickUp() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockRollingStoneInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockRollingStoneInteractor) GetConfig() domain.RollingStoneConfig {
	return _m.Called().Get(0).(domain.RollingStoneConfig)
}

// Hint モック
func (_m *MockRollingStoneInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockRollingStoneInteractor) ActionLog() string { return _m.Called().String(0) }

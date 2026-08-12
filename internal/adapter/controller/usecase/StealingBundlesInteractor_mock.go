//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockStealingBundlesInteractor スティーリングバンドルインタラクターモック
type MockStealingBundlesInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockStealingBundlesInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockStealingBundlesInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockStealingBundlesInteractor) ResetWithConfig(cfg domain.StealingBundlesConfig) string {
	return _m.Called(cfg).String(0)
}

// Take モック
func (_m *MockStealingBundlesInteractor) Take(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

// Steal モック
func (_m *MockStealingBundlesInteractor) Steal(cardIndex, victimIdx int) string {
	return _m.Called(cardIndex, victimIdx).String(0)
}

// Trail モック
func (_m *MockStealingBundlesInteractor) Trail(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

// GiveUp モック
func (_m *MockStealingBundlesInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockStealingBundlesInteractor) GetConfig() domain.StealingBundlesConfig {
	return _m.Called().Get(0).(domain.StealingBundlesConfig)
}

// Hint モック
func (_m *MockStealingBundlesInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockStealingBundlesInteractor) ActionLog() string { return _m.Called().String(0) }

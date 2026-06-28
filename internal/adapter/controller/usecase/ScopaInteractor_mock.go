//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockScopaInteractor スコパインタラクターモック。
type MockScopaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockScopaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockScopaInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockScopaInteractor) Play(handIdx int, tableIdxs []int) string {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockScopaInteractor) GetConfig() domain.ScopaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ScopaConfig)
}

// ResetWithConfig モック
func (_m *MockScopaInteractor) ResetWithConfig(config domain.ScopaConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockScopaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockScopaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockScopaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockScoponeInteractor スコポーネインタラクターモック。
type MockScoponeInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockScoponeInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockScoponeInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockScoponeInteractor) Play(handIdx int, tableIdxs []int) string {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockScoponeInteractor) GetConfig() domain.ScoponeConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ScoponeConfig)
}

// ResetWithConfig モック
func (_m *MockScoponeInteractor) ResetWithConfig(config domain.ScoponeConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockScoponeInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockScoponeInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

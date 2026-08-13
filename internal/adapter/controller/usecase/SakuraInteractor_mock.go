//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSakuraInteractor はさくら (肥後花) のインタラクターモック。
type MockSakuraInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSakuraInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSakuraInteractor) ResetWithConfig(cfg domain.SakuraConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSakuraInteractor) Play(handIdx, fieldIdx int) string {
	ret := _m.Called(handIdx, fieldIdx)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockSakuraInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSakuraInteractor) GetConfig() domain.SakuraConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SakuraConfig)
}

// Hint モック
func (_m *MockSakuraInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSakuraInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSakuraInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

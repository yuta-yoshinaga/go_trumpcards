//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockThreeCardBragInteractor スリーカード・ブラグのインタラクターモック
type MockThreeCardBragInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockThreeCardBragInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockThreeCardBragInteractor) ResetWithConfig(cfg domain.ThreeCardBragConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// See モック
func (_m *MockThreeCardBragInteractor) See() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Bet モック
func (_m *MockThreeCardBragInteractor) Bet() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Raise モック
func (_m *MockThreeCardBragInteractor) Raise(newStake int) string {
	ret := _m.Called(newStake)
	return ret.Get(0).(string)
}

// Fold モック
func (_m *MockThreeCardBragInteractor) Fold() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Show モック
func (_m *MockThreeCardBragInteractor) Show() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockThreeCardBragInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockThreeCardBragInteractor) GetConfig() domain.ThreeCardBragConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ThreeCardBragConfig)
}

// Hint モック
func (_m *MockThreeCardBragInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockThreeCardBragInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockThreeCardBragInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

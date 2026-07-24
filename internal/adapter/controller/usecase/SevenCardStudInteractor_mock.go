//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSevenCardStudInteractor セブンカードスタッドインタラクターモック
type MockSevenCardStudInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSevenCardStudInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSevenCardStudInteractor) ResetWithConfig(cfg domain.SevenCardStudConfig, profileData []byte) string {
	ret := _m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockSevenCardStudInteractor) Action(action int, amount int, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSevenCardStudInteractor) GetConfig() domain.SevenCardStudConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SevenCardStudConfig)
}

// Rebuy モック
func (_m *MockSevenCardStudInteractor) Rebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipRebuy モック
func (_m *MockSevenCardStudInteractor) SkipRebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Addon モック
func (_m *MockSevenCardStudInteractor) Addon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipAddon モック
func (_m *MockSevenCardStudInteractor) SkipAddon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Muck モック
func (_m *MockSevenCardStudInteractor) Muck() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ShowHand モック
func (_m *MockSevenCardStudInteractor) ShowHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSevenCardStudInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockSevenCardStudInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSevenCardStudInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

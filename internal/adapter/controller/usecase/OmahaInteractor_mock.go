//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOmahaInteractor オマハホールデムインタラクターモック
type MockOmahaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockOmahaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockOmahaInteractor) ResetWithConfig(cfg domain.OmahaConfig, profileData []byte) string {
	ret := _m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockOmahaInteractor) Action(action int, amount int, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockOmahaInteractor) GetConfig() domain.OmahaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.OmahaConfig)
}

// Rebuy モック
func (_m *MockOmahaInteractor) Rebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipRebuy モック
func (_m *MockOmahaInteractor) SkipRebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Addon モック
func (_m *MockOmahaInteractor) Addon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipAddon モック
func (_m *MockOmahaInteractor) SkipAddon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Muck モック
func (_m *MockOmahaInteractor) Muck() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ShowHand モック
func (_m *MockOmahaInteractor) ShowHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockOmahaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

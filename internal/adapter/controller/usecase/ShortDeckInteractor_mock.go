package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShortDeckInteractor ショートデックホールデムインタラクターモック
type MockShortDeckInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockShortDeckInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockShortDeckInteractor) ResetWithConfig(cfg domain.ShortDeckConfig, profileData []byte) string {
	ret := _m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockShortDeckInteractor) Action(action int, amount int, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockShortDeckInteractor) GetConfig() domain.ShortDeckConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ShortDeckConfig)
}

// Rebuy モック
func (_m *MockShortDeckInteractor) Rebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipRebuy モック
func (_m *MockShortDeckInteractor) SkipRebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Addon モック
func (_m *MockShortDeckInteractor) Addon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipAddon モック
func (_m *MockShortDeckInteractor) SkipAddon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Muck モック
func (_m *MockShortDeckInteractor) Muck() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ShowHand モック
func (_m *MockShortDeckInteractor) ShowHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockShortDeckInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHoldemInteractor テキサスホールデムインタラクターモック
type MockHoldemInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockHoldemInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockHoldemInteractor) ResetWithConfig(cfg domain.HoldemConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockHoldemInteractor) Action(action int, amount int) string {
	ret := _m.Called(action, amount)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockHoldemInteractor) GetConfig() domain.HoldemConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.HoldemConfig)
}

// Rebuy モック
func (_m *MockHoldemInteractor) Rebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipRebuy モック
func (_m *MockHoldemInteractor) SkipRebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Addon モック
func (_m *MockHoldemInteractor) Addon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipAddon モック
func (_m *MockHoldemInteractor) SkipAddon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

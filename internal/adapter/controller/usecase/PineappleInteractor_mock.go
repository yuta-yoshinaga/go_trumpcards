//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPineappleInteractor パイナップルポーカーインタラクターモック
type MockPineappleInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPineappleInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockPineappleInteractor) ResetWithConfig(cfg domain.PineappleConfig, profileData []byte) string {
	ret := _m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockPineappleInteractor) Action(action int, amount int, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockPineappleInteractor) Discard(cardIdx int) string {
	ret := _m.Called(cardIdx)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPineappleInteractor) GetConfig() domain.PineappleConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PineappleConfig)
}

// Rebuy モック
func (_m *MockPineappleInteractor) Rebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipRebuy モック
func (_m *MockPineappleInteractor) SkipRebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Addon モック
func (_m *MockPineappleInteractor) Addon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipAddon モック
func (_m *MockPineappleInteractor) SkipAddon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Muck モック
func (_m *MockPineappleInteractor) Muck() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ShowHand モック
func (_m *MockPineappleInteractor) ShowHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPineappleInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

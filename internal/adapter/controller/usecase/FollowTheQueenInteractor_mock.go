//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFollowTheQueenInteractor フォロー・ザ・クイーンインタラクターモック
type MockFollowTheQueenInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockFollowTheQueenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockFollowTheQueenInteractor) ResetWithConfig(cfg domain.FollowTheQueenConfig, profileData []byte) string {
	ret := _m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockFollowTheQueenInteractor) Action(action int, amount int, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockFollowTheQueenInteractor) GetConfig() domain.FollowTheQueenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.FollowTheQueenConfig)
}

// Rebuy モック
func (_m *MockFollowTheQueenInteractor) Rebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipRebuy モック
func (_m *MockFollowTheQueenInteractor) SkipRebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Addon モック
func (_m *MockFollowTheQueenInteractor) Addon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipAddon モック
func (_m *MockFollowTheQueenInteractor) SkipAddon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Muck モック
func (_m *MockFollowTheQueenInteractor) Muck() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ShowHand モック
func (_m *MockFollowTheQueenInteractor) ShowHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockFollowTheQueenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockFollowTheQueenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockFollowTheQueenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

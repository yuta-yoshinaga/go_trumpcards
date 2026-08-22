//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDramahaInteractor ドラマハホールデムインタラクターモック
type MockDramahaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockDramahaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockDramahaInteractor) ResetWithConfig(cfg domain.DramahaConfig, profileData []byte) string {
	ret := _m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockDramahaInteractor) Action(action int, amount int, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// Draw モック
func (_m *MockDramahaInteractor) Draw(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockDramahaInteractor) GetConfig() domain.DramahaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DramahaConfig)
}

// Rebuy モック
func (_m *MockDramahaInteractor) Rebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipRebuy モック
func (_m *MockDramahaInteractor) SkipRebuy() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Addon モック
func (_m *MockDramahaInteractor) Addon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SkipAddon モック
func (_m *MockDramahaInteractor) SkipAddon() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Muck モック
func (_m *MockDramahaInteractor) Muck() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ShowHand モック
func (_m *MockDramahaInteractor) ShowHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockDramahaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockDramahaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

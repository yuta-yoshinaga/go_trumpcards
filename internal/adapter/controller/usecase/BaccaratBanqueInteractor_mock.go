//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBaccaratBanqueInteractor はバカラ・バンクのインタラクターモック。
type MockBaccaratBanqueInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockBaccaratBanqueInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockBaccaratBanqueInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBaccaratBanqueInteractor) ResetWithConfig(config domain.BaccaratBanqueConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Draw モック
func (_m *MockBaccaratBanqueInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Stand モック
func (_m *MockBaccaratBanqueInteractor) Stand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextCoup モック
func (_m *MockBaccaratBanqueInteractor) NextCoup() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Retire モック
func (_m *MockBaccaratBanqueInteractor) Retire() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBaccaratBanqueInteractor) GetConfig() domain.BaccaratBanqueConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BaccaratBanqueConfig)
}

// Hint モック
func (_m *MockBaccaratBanqueInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBaccaratBanqueInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

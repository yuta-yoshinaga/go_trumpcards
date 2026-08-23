//go:build test && (!js || !wasm || classic)

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBrusquembilleInteractor ブリュスカンビーユインタラクターモック
type MockBrusquembilleInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBrusquembilleInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBrusquembilleInteractor) ResetWithConfig(cfg domain.BrusquembilleConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBrusquembilleInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockBrusquembilleInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBrusquembilleInteractor) GetConfig() domain.BrusquembilleConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BrusquembilleConfig)
}

// Hint モック
func (_m *MockBrusquembilleInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBrusquembilleInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBrusquembilleInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

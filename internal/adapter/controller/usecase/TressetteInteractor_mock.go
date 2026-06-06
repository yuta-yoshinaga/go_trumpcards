//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTressetteInteractor トレセッテのインタラクターモック
type MockTressetteInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTressetteInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTressetteInteractor) ResetWithConfig(cfg domain.TressetteConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTressetteInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockTressetteInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTressetteInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTressetteInteractor) GetConfig() domain.TressetteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TressetteConfig)
}

// Hint モック
func (_m *MockTressetteInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTressetteInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTressetteInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

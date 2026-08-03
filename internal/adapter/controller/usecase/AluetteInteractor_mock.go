//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAluetteInteractor アリュエットのインタラクターモック
type MockAluetteInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockAluetteInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockAluetteInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockAluetteInteractor) ResetWithConfig(cfg domain.AluetteConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockAluetteInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockAluetteInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockAluetteInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockAluetteInteractor) GetConfig() domain.AluetteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.AluetteConfig)
}

// Hint モック
func (_m *MockAluetteInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockAluetteInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

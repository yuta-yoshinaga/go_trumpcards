//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBouillotteInteractor はブイヨット (Bouillotte) のインタラクターモック。
type MockBouillotteInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBouillotteInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBouillotteInteractor) ResetWithConfig(cfg domain.BouillotteConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bet モック
func (_m *MockBouillotteInteractor) Bet(action string) string {
	ret := _m.Called(action)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBouillotteInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBouillotteInteractor) GetConfig() domain.BouillotteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BouillotteConfig)
}

// Hint モック
func (_m *MockBouillotteInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBouillotteInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBouillotteInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

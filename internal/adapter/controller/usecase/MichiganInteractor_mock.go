//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMichiganInteractor はミシガン (Michigan) のインタラクターモック。
type MockMichiganInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockMichiganInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockMichiganInteractor) ResetWithConfig(cfg domain.MichiganConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bet モック
func (_m *MockMichiganInteractor) Bet(bets []int) string {
	ret := _m.Called(bets)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockMichiganInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockMichiganInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockMichiganInteractor) GetConfig() domain.MichiganConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MichiganConfig)
}

// Hint モック
func (_m *MockMichiganInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockMichiganInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMichiganInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

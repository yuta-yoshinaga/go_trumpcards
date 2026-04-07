//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPokerInteractor ポーカーインタラクターモック
type MockPokerInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPokerInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPokerInteractor) GetConfig() domain.PokerConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PokerConfig)
}

// ResetWithConfig モック
func (_m *MockPokerInteractor) ResetWithConfig(cfg domain.PokerConfig, profileData []byte) string {
	ret := _m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockPokerInteractor) Action(action int, amount int, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// Exchange モック
func (_m *MockPokerInteractor) Exchange(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Stand モック
func (_m *MockPokerInteractor) Stand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Odds モック
func (_m *MockPokerInteractor) Odds(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPokerInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPokerInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

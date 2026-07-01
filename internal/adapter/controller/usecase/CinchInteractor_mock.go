//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCinchInteractor はチンチ (Cinch) のインタラクターモック。
type MockCinchInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCinchInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCinchInteractor) ResetWithConfig(cfg domain.CinchConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockCinchInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// NameTrump モック
func (_m *MockCinchInteractor) NameTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCinchInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCinchInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCinchInteractor) GetConfig() domain.CinchConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CinchConfig)
}

// Hint モック
func (_m *MockCinchInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCinchInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCinchInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

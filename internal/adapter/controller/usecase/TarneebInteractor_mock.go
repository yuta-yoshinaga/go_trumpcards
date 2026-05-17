//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTarneebInteractor Tarneeb インタラクターモック
type MockTarneebInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTarneebInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTarneebInteractor) ResetWithConfig(cfg domain.TarneebConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockTarneebInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// DeclareTrump モック
func (_m *MockTarneebInteractor) DeclareTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTarneebInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockTarneebInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTarneebInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTarneebInteractor) GetConfig() domain.TarneebConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TarneebConfig)
}

// Hint モック
func (_m *MockTarneebInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTarneebInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTarneebInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

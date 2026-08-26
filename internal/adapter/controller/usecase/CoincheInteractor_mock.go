//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCoincheInteractor コワンシュインタラクターモック
type MockCoincheInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCoincheInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCoincheInteractor) ResetWithConfig(cfg domain.CoincheConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockCoincheInteractor) Bid(points, suit int) string {
	ret := _m.Called(points, suit)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockCoincheInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Coinche モック
func (_m *MockCoincheInteractor) Coinche() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Surcoinche モック
func (_m *MockCoincheInteractor) Surcoinche() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// DeclineDouble モック
func (_m *MockCoincheInteractor) DeclineDouble() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCoincheInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockCoincheInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCoincheInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCoincheInteractor) GetConfig() domain.CoincheConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CoincheConfig)
}

// Hint モック
func (_m *MockCoincheInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCoincheInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCoincheInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

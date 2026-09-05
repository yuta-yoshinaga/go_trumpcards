//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBinokelInteractor ビノクルインタラクターモック
type MockBinokelInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBinokelInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBinokelInteractor) ResetWithConfig(cfg domain.BinokelConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockBinokelInteractor) Bid(amount int) string {
	ret := _m.Called(amount)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockBinokelInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// DiscardToDabb モック
func (_m *MockBinokelInteractor) DiscardToDabb(cardIndices []int) string {
	ret := _m.Called(cardIndices)
	return ret.Get(0).(string)
}

// CallTrump モック
func (_m *MockBinokelInteractor) CallTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// ConfirmMelds モック
func (_m *MockBinokelInteractor) ConfirmMelds() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBinokelInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockBinokelInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBinokelInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBinokelInteractor) GetConfig() domain.BinokelConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BinokelConfig)
}

// Hint モック
func (_m *MockBinokelInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBinokelInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBinokelInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

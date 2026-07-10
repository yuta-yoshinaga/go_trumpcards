//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockUltiInteractor ウルティ (Ulti) のインタラクターモック
type MockUltiInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockUltiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockUltiInteractor) ResetWithConfig(cfg domain.UltiConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockUltiInteractor) Bid(contract domain.UltiContract, trumpSuit int) string {
	ret := _m.Called(contract, trumpSuit)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockUltiInteractor) Discard(cardIndices []int) string {
	ret := _m.Called(cardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockUltiInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockUltiInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockUltiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockUltiInteractor) GetConfig() domain.UltiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.UltiConfig)
}

// Hint モック
func (_m *MockUltiInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockUltiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockUltiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

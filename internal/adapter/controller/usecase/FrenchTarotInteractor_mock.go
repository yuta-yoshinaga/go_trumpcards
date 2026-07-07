//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFrenchTarotInteractor フレンチタロット (French Tarot) のインタラクターモック
type MockFrenchTarotInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockFrenchTarotInteractor) Reset() string {
	return _m.Called().Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockFrenchTarotInteractor) ResetWithConfig(cfg domain.FrenchTarotConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Bid モック
func (_m *MockFrenchTarotInteractor) Bid(bid domain.FrenchTarotBid) string {
	return _m.Called(bid).Get(0).(string)
}

// Pass モック
func (_m *MockFrenchTarotInteractor) Pass() string {
	return _m.Called().Get(0).(string)
}

// Discard モック
func (_m *MockFrenchTarotInteractor) Discard(cardIndices []int) string {
	return _m.Called(cardIndices).Get(0).(string)
}

// Play モック
func (_m *MockFrenchTarotInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextTrick モック
func (_m *MockFrenchTarotInteractor) NextTrick() string {
	return _m.Called().Get(0).(string)
}

// NextRound モック
func (_m *MockFrenchTarotInteractor) NextRound() string {
	return _m.Called().Get(0).(string)
}

// GetConfig モック
func (_m *MockFrenchTarotInteractor) GetConfig() domain.FrenchTarotConfig {
	return _m.Called().Get(0).(domain.FrenchTarotConfig)
}

// Hint モック
func (_m *MockFrenchTarotInteractor) Hint() string {
	return _m.Called().Get(0).(string)
}

// ActionLog モック
func (_m *MockFrenchTarotInteractor) ActionLog() string {
	return _m.Called().Get(0).(string)
}

// Snapshot モック
func (_m *MockFrenchTarotInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

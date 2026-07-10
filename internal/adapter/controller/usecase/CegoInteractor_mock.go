//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCegoInteractor チェゴ (Cego) のインタラクターモック
type MockCegoInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCegoInteractor) Reset() string {
	return _m.Called().Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCegoInteractor) ResetWithConfig(cfg domain.CegoConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Bid モック
func (_m *MockCegoInteractor) Bid(bid domain.CegoBid) string {
	return _m.Called(bid).Get(0).(string)
}

// Pass モック
func (_m *MockCegoInteractor) Pass() string {
	return _m.Called().Get(0).(string)
}

// ChooseContract モック
func (_m *MockCegoInteractor) ChooseContract(ct domain.CegoContract) string {
	return _m.Called(ct).Get(0).(string)
}

// Discard モック
func (_m *MockCegoInteractor) Discard(keepIndices []int) string {
	return _m.Called(keepIndices).Get(0).(string)
}

// Play モック
func (_m *MockCegoInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextTrick モック
func (_m *MockCegoInteractor) NextTrick() string {
	return _m.Called().Get(0).(string)
}

// NextRound モック
func (_m *MockCegoInteractor) NextRound() string {
	return _m.Called().Get(0).(string)
}

// GetConfig モック
func (_m *MockCegoInteractor) GetConfig() domain.CegoConfig {
	return _m.Called().Get(0).(domain.CegoConfig)
}

// Hint モック
func (_m *MockCegoInteractor) Hint() string {
	return _m.Called().Get(0).(string)
}

// ActionLog モック
func (_m *MockCegoInteractor) ActionLog() string {
	return _m.Called().Get(0).(string)
}

// Snapshot モック
func (_m *MockCegoInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

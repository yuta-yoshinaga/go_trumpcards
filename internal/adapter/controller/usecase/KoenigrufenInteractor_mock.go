//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKoenigrufenInteractor ケーニッヒルーフェン (Königrufen) のインタラクターモック
type MockKoenigrufenInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockKoenigrufenInteractor) Reset() string {
	return _m.Called().Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockKoenigrufenInteractor) ResetWithConfig(cfg domain.KoenigrufenConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Bid モック
func (_m *MockKoenigrufenInteractor) Bid(bid domain.KoenigrufenBid) string {
	return _m.Called(bid).Get(0).(string)
}

// Pass モック
func (_m *MockKoenigrufenInteractor) Pass() string {
	return _m.Called().Get(0).(string)
}

// CallKing モック
func (_m *MockKoenigrufenInteractor) CallKing(suit int) string {
	return _m.Called(suit).Get(0).(string)
}

// Discard モック
func (_m *MockKoenigrufenInteractor) Discard(cardIndices []int) string {
	return _m.Called(cardIndices).Get(0).(string)
}

// Play モック
func (_m *MockKoenigrufenInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextTrick モック
func (_m *MockKoenigrufenInteractor) NextTrick() string {
	return _m.Called().Get(0).(string)
}

// NextRound モック
func (_m *MockKoenigrufenInteractor) NextRound() string {
	return _m.Called().Get(0).(string)
}

// GetConfig モック
func (_m *MockKoenigrufenInteractor) GetConfig() domain.KoenigrufenConfig {
	return _m.Called().Get(0).(domain.KoenigrufenConfig)
}

// Hint モック
func (_m *MockKoenigrufenInteractor) Hint() string {
	return _m.Called().Get(0).(string)
}

// ActionLog モック
func (_m *MockKoenigrufenInteractor) ActionLog() string {
	return _m.Called().Get(0).(string)
}

// Snapshot モック
func (_m *MockKoenigrufenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

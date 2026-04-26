//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSkatInteractor mocks the Skat interactor.
type MockSkatInteractor struct {
	mock.Mock
}

// Reset mock.
func (_m *MockSkatInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig mock.
func (_m *MockSkatInteractor) ResetWithConfig(cfg domain.SkatConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Bid mock.
func (_m *MockSkatInteractor) Bid(accept bool) string {
	return _m.Called(accept).Get(0).(string)
}

// PickSkat mock.
func (_m *MockSkatInteractor) PickSkat(pickup bool) string {
	return _m.Called(pickup).Get(0).(string)
}

// Discard mock.
func (_m *MockSkatInteractor) Discard(idxA, idxB int) string {
	return _m.Called(idxA, idxB).Get(0).(string)
}

// DeclareGame mock.
func (_m *MockSkatInteractor) DeclareGame(gameType domain.SkatGameType, trumpSuit int) string {
	return _m.Called(gameType, trumpSuit).Get(0).(string)
}

// Play mock.
func (_m *MockSkatInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextTrick mock.
func (_m *MockSkatInteractor) NextTrick() string { return _m.Called().Get(0).(string) }

// NextRound mock.
func (_m *MockSkatInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GetConfig mock.
func (_m *MockSkatInteractor) GetConfig() domain.SkatConfig {
	return _m.Called().Get(0).(domain.SkatConfig)
}

// Hint mock.
func (_m *MockSkatInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog mock.
func (_m *MockSkatInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot mock.
func (_m *MockSkatInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

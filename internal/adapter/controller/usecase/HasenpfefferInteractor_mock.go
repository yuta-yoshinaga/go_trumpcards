//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHasenpfefferInteractor ハーゼンプフェファーインタラクターモック
type MockHasenpfefferInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockHasenpfefferInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockHasenpfefferInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockHasenpfefferInteractor) ResetWithConfig(cfg domain.HasenpfefferConfig) string {
	return _m.Called(cfg).String(0)
}

// Bid モック
func (_m *MockHasenpfefferInteractor) Bid(bid int) string { return _m.Called(bid).String(0) }

// Discard モック
func (_m *MockHasenpfefferInteractor) Discard(cardIndex, suit int) string {
	return _m.Called(cardIndex, suit).String(0)
}

// Play モック
func (_m *MockHasenpfefferInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

// NextHand モック
func (_m *MockHasenpfefferInteractor) NextHand() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockHasenpfefferInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockHasenpfefferInteractor) GetConfig() domain.HasenpfefferConfig {
	return _m.Called().Get(0).(domain.HasenpfefferConfig)
}

// Hint モック
func (_m *MockHasenpfefferInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockHasenpfefferInteractor) ActionLog() string { return _m.Called().String(0) }

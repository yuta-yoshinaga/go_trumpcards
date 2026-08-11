//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShelemInteractor シェレムインタラクターモック
type MockShelemInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockShelemInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockShelemInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockShelemInteractor) ResetWithConfig(cfg domain.ShelemConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Bid モック
func (_m *MockShelemInteractor) Bid(bid int) string { return _m.Called(bid).Get(0).(string) }

// BidShelem モック
func (_m *MockShelemInteractor) BidShelem() string { return _m.Called().Get(0).(string) }

// Pass モック
func (_m *MockShelemInteractor) Pass() string { return _m.Called().Get(0).(string) }

// Discard モック
func (_m *MockShelemInteractor) Discard(indices []int, suit int) string {
	return _m.Called(indices, suit).Get(0).(string)
}

// Play モック
func (_m *MockShelemInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockShelemInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockShelemInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockShelemInteractor) GetConfig() domain.ShelemConfig {
	return _m.Called().Get(0).(domain.ShelemConfig)
}

// Hint モック
func (_m *MockShelemInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockShelemInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEstimationInteractor エスティメーションインタラクターモック
type MockEstimationInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockEstimationInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockEstimationInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockEstimationInteractor) ResetWithConfig(cfg domain.EstimationConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// SelectTrump モック
func (_m *MockEstimationInteractor) SelectTrump(suit int) string {
	return _m.Called(suit).Get(0).(string)
}

// Bid モック
func (_m *MockEstimationInteractor) Bid(bid int) string { return _m.Called(bid).Get(0).(string) }

// Play モック
func (_m *MockEstimationInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockEstimationInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockEstimationInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockEstimationInteractor) GetConfig() domain.EstimationConfig {
	return _m.Called().Get(0).(domain.EstimationConfig)
}

// Hint モック
func (_m *MockEstimationInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockEstimationInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

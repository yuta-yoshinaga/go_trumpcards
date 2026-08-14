//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockIsraeliWhistInteractor イスラエリホイストインタラクターモック
type MockIsraeliWhistInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockIsraeliWhistInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockIsraeliWhistInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockIsraeliWhistInteractor) ResetWithConfig(cfg domain.IsraeliWhistConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// AuctionBid モック
func (_m *MockIsraeliWhistInteractor) AuctionBid(bid, suit int) string {
	return _m.Called(bid, suit).Get(0).(string)
}

// AuctionPass モック
func (_m *MockIsraeliWhistInteractor) AuctionPass() string { return _m.Called().Get(0).(string) }

// Bid モック
func (_m *MockIsraeliWhistInteractor) Bid(bid int) string { return _m.Called(bid).Get(0).(string) }

// Play モック
func (_m *MockIsraeliWhistInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockIsraeliWhistInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockIsraeliWhistInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockIsraeliWhistInteractor) GetConfig() domain.IsraeliWhistConfig {
	return _m.Called().Get(0).(domain.IsraeliWhistConfig)
}

// Hint モック
func (_m *MockIsraeliWhistInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockIsraeliWhistInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBidEuchreInteractor モック
type MockBidEuchreInteractor struct {
	mock.Mock
}

func (_m *MockBidEuchreInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockBidEuchreInteractor) ResetWithConfig(cfg domain.BidEuchreConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockBidEuchreInteractor) Bid(value int) string {
	return _m.Called(value).String(0)
}

func (_m *MockBidEuchreInteractor) PassBid() string {
	return _m.Called().String(0)
}

func (_m *MockBidEuchreInteractor) ChooseTrump(t int) string {
	return _m.Called(t).String(0)
}

func (_m *MockBidEuchreInteractor) PlayCard(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockBidEuchreInteractor) NextHand() string {
	return _m.Called().String(0)
}

func (_m *MockBidEuchreInteractor) GetConfig() domain.BidEuchreConfig {
	return _m.Called().Get(0).(domain.BidEuchreConfig)
}

func (_m *MockBidEuchreInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockBidEuchreInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

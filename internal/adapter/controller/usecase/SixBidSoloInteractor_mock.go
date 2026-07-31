//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSixBidSoloInteractor モック
type MockSixBidSoloInteractor struct {
	mock.Mock
}

func (_m *MockSixBidSoloInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockSixBidSoloInteractor) ResetWithConfig(cfg domain.SixBidSoloConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockSixBidSoloInteractor) Bid(kind int) string {
	return _m.Called(kind).String(0)
}

func (_m *MockSixBidSoloInteractor) PassBid() string {
	return _m.Called().String(0)
}

func (_m *MockSixBidSoloInteractor) Declare(suit, calledSuit, calledValue int) string {
	return _m.Called(suit, calledSuit, calledValue).String(0)
}

func (_m *MockSixBidSoloInteractor) PlayCard(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockSixBidSoloInteractor) NextHand() string {
	return _m.Called().String(0)
}

func (_m *MockSixBidSoloInteractor) GetConfig() domain.SixBidSoloConfig {
	return _m.Called().Get(0).(domain.SixBidSoloConfig)
}

func (_m *MockSixBidSoloInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockSixBidSoloInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

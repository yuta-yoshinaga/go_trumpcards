//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBostonInteractor モック
type MockBostonInteractor struct {
	mock.Mock
}

func (_m *MockBostonInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockBostonInteractor) ResetWithConfig(cfg domain.BostonConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockBostonInteractor) Bid(level domain.BostonBidLevel, suit int) string {
	return _m.Called(level, suit).String(0)
}

func (_m *MockBostonInteractor) PassBid() string {
	return _m.Called().String(0)
}

func (_m *MockBostonInteractor) CallPartner(partner int) string {
	return _m.Called(partner).String(0)
}

func (_m *MockBostonInteractor) PlayCard(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockBostonInteractor) NextHand() string {
	return _m.Called().String(0)
}

func (_m *MockBostonInteractor) GetConfig() domain.BostonConfig {
	return _m.Called().Get(0).(domain.BostonConfig)
}

func (_m *MockBostonInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockBostonInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

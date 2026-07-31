//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockVintInteractor モック
type MockVintInteractor struct {
	mock.Mock
}

func (_m *MockVintInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockVintInteractor) ResetWithConfig(cfg domain.VintConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockVintInteractor) Bid(level, denom int) string {
	return _m.Called(level, denom).String(0)
}

func (_m *MockVintInteractor) PassBid() string {
	return _m.Called().String(0)
}

func (_m *MockVintInteractor) PlayCard(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockVintInteractor) NextHand() string {
	return _m.Called().String(0)
}

func (_m *MockVintInteractor) GetConfig() domain.VintConfig {
	return _m.Called().Get(0).(domain.VintConfig)
}

func (_m *MockVintInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockVintInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShengJiInteractor モック
type MockShengJiInteractor struct {
	mock.Mock
}

func (_m *MockShengJiInteractor) Reset() string { return _m.Called().String(0) }

func (_m *MockShengJiInteractor) ResetWithConfig(cfg domain.ShengJiConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockShengJiInteractor) Declare(suit int) string {
	return _m.Called(suit).String(0)
}

func (_m *MockShengJiInteractor) BuryKitty(idxs []int) string {
	return _m.Called(idxs).String(0)
}

func (_m *MockShengJiInteractor) Play(idxs []int) string {
	return _m.Called(idxs).String(0)
}

func (_m *MockShengJiInteractor) NextHand() string { return _m.Called().String(0) }

func (_m *MockShengJiInteractor) GetConfig() domain.ShengJiConfig {
	return _m.Called().Get(0).(domain.ShengJiConfig)
}

func (_m *MockShengJiInteractor) ActionLog() string { return _m.Called().String(0) }

// Snapshot モック
func (_m *MockShengJiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

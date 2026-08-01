//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGuandanInteractor モック
type MockGuandanInteractor struct {
	mock.Mock
}

func (_m *MockGuandanInteractor) Reset() string { return _m.Called().String(0) }

func (_m *MockGuandanInteractor) ResetWithConfig(cfg domain.GuandanConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockGuandanInteractor) PlayCards(idxs []int) string {
	return _m.Called(idxs).String(0)
}

func (_m *MockGuandanInteractor) Pass() string { return _m.Called().String(0) }

func (_m *MockGuandanInteractor) ReturnTribute(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockGuandanInteractor) NextHand() string { return _m.Called().String(0) }

func (_m *MockGuandanInteractor) GetConfig() domain.GuandanConfig {
	return _m.Called().Get(0).(domain.GuandanConfig)
}

func (_m *MockGuandanInteractor) ActionLog() string { return _m.Called().String(0) }

// Snapshot モック
func (_m *MockGuandanInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

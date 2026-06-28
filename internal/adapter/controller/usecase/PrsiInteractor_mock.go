//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPrsiInteractor モック
type MockPrsiInteractor struct {
	mock.Mock
}

func (_m *MockPrsiInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockPrsiInteractor) ResetWithConfig(cfg domain.PrsiConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockPrsiInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockPrsiInteractor) Draw() string {
	return _m.Called().String(0)
}

func (_m *MockPrsiInteractor) GetConfig() domain.PrsiConfig {
	return _m.Called().Get(0).(domain.PrsiConfig)
}

func (_m *MockPrsiInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockPrsiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

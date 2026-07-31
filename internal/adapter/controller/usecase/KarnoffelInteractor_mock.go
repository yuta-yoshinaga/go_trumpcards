//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKarnoffelInteractor モック
type MockKarnoffelInteractor struct {
	mock.Mock
}

func (_m *MockKarnoffelInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockKarnoffelInteractor) ResetWithConfig(cfg domain.KarnoffelConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockKarnoffelInteractor) PlayCard(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockKarnoffelInteractor) NextHand() string {
	return _m.Called().String(0)
}

func (_m *MockKarnoffelInteractor) GetConfig() domain.KarnoffelConfig {
	return _m.Called().Get(0).(domain.KarnoffelConfig)
}

func (_m *MockKarnoffelInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockKarnoffelInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSjavsInteractor シャウス インタラクターモック
type MockSjavsInteractor struct {
	mock.Mock
}

func (_m *MockSjavsInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockSjavsInteractor) ResetWithConfig(cfg domain.SjavsConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockSjavsInteractor) Bid(length int) string { return _m.Called(length).Get(0).(string) }

func (_m *MockSjavsInteractor) Play(handIdx int) string { return _m.Called(handIdx).Get(0).(string) }

func (_m *MockSjavsInteractor) NextHand() string { return _m.Called().Get(0).(string) }

func (_m *MockSjavsInteractor) GetConfig() domain.SjavsConfig {
	return _m.Called().Get(0).(domain.SjavsConfig)
}

func (_m *MockSjavsInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockSjavsInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockSjavsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

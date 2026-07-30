//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSkitgubbeInteractor シートグッベ インタラクターモック
type MockSkitgubbeInteractor struct {
	mock.Mock
}

func (_m *MockSkitgubbeInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockSkitgubbeInteractor) ResetWithConfig(cfg domain.SkitgubbeConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockSkitgubbeInteractor) Play(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockSkitgubbeInteractor) PickUp() string { return _m.Called().Get(0).(string) }

func (_m *MockSkitgubbeInteractor) GetConfig() domain.SkitgubbeConfig {
	return _m.Called().Get(0).(domain.SkitgubbeConfig)
}

func (_m *MockSkitgubbeInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockSkitgubbeInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockSkitgubbeInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

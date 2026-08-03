//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockChineseTenInteractor 撿紅點 インタラクターモック
type MockChineseTenInteractor struct {
	mock.Mock
}

func (_m *MockChineseTenInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockChineseTenInteractor) ResetWithConfig(cfg domain.ChineseTenConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockChineseTenInteractor) Play(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockChineseTenInteractor) Select(layoutIdx int) string {
	return _m.Called(layoutIdx).Get(0).(string)
}

func (_m *MockChineseTenInteractor) GetConfig() domain.ChineseTenConfig {
	return _m.Called().Get(0).(domain.ChineseTenConfig)
}

func (_m *MockChineseTenInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockChineseTenInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockChineseTenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

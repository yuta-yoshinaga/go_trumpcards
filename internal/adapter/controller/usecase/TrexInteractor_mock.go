//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrexInteractor トリックス インタラクターモック
type MockTrexInteractor struct {
	mock.Mock
}

func (_m *MockTrexInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockTrexInteractor) ResetWithConfig(cfg domain.TrexConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockTrexInteractor) Choose(contract int) string {
	return _m.Called(contract).Get(0).(string)
}

func (_m *MockTrexInteractor) Play(handIdx int) string { return _m.Called(handIdx).Get(0).(string) }

func (_m *MockTrexInteractor) Pass() string { return _m.Called().Get(0).(string) }

func (_m *MockTrexInteractor) NextDeal() string { return _m.Called().Get(0).(string) }

func (_m *MockTrexInteractor) GetConfig() domain.TrexConfig {
	return _m.Called().Get(0).(domain.TrexConfig)
}

func (_m *MockTrexInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockTrexInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockTrexInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

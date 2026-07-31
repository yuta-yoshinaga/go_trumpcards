//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPochInteractor ポッホ インタラクターモック
type MockPochInteractor struct {
	mock.Mock
}

func (_m *MockPochInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockPochInteractor) ResetWithConfig(cfg domain.PochConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockPochInteractor) Bet() string { return _m.Called().Get(0).(string) }

func (_m *MockPochInteractor) Fold() string { return _m.Called().Get(0).(string) }

func (_m *MockPochInteractor) Play(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockPochInteractor) NextDeal() string { return _m.Called().Get(0).(string) }

func (_m *MockPochInteractor) GetConfig() domain.PochConfig {
	return _m.Called().Get(0).(domain.PochConfig)
}

func (_m *MockPochInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockPochInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockPochInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

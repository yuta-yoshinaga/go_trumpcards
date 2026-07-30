//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLaughAndLieDownInteractor ラフ・アンド・ライダウン インタラクターモック
type MockLaughAndLieDownInteractor struct {
	mock.Mock
}

func (_m *MockLaughAndLieDownInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockLaughAndLieDownInteractor) ResetWithConfig(cfg domain.LaughAndLieDownConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockLaughAndLieDownInteractor) Play(handIdx, takeCount int) string {
	return _m.Called(handIdx, takeCount).Get(0).(string)
}

func (_m *MockLaughAndLieDownInteractor) GetConfig() domain.LaughAndLieDownConfig {
	return _m.Called().Get(0).(domain.LaughAndLieDownConfig)
}

func (_m *MockLaughAndLieDownInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockLaughAndLieDownInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockLaughAndLieDownInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
